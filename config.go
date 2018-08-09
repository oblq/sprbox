package sprbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

// struct field flags
const (
	// sftEnv value can be in json format, it will override also the default value
	sftEnv = "env"

	// set the default value
	sftDefault = "default"

	// return error if missing value
	sftRequired = "required"
)

// files type regexp
const (
	regexYAML = `(?i)(.y(|a)ml)`
	regexTOML = `(?i)(.toml)`
	regexJSON = `(?i)(.json)`

	regexExt = `(?i)(.y(|a)ml|.toml|.json)` // `(?i)(\..{3,4})` //
)

// FILE SEARCH ---------------------------------------------------------------------------------------------------------

// walkConfigPath look for a file matching the passed regex skipping sub-directories.
func walkConfigPath(configPath string, regex *regexp.Regexp) (matchedFile string) {
	err := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		// nil if the path does not exist
		if info == nil {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() != filepath.Base(configPath) {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if regex.MatchString(info.Name()) {
			matchedFile = path
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	return
}

// configFilesByEnv will search for the given file in the given path
// returning all the eligible files (eg.: <path>/myConfig.yaml and <path>/myConfig.<environment>.yaml)
//
// 'filePath' can also be passed without file extension,
// searchFilesByEnv is agnostic and will match any
// supported extension in that case.
//
// The 'file' name will be searched as (in that order):
//  - '<path>/<file>(.* || <the_provided_extension>)'
//  - '<path>/<file>.<environment>(.* || <the_provided_extension>)'
//
// The latest found files will override previous.
func configFilesByEnv(files ...string) (foundFiles []string) {
	for _, file := range files {
		configPath, fileName := filepath.Split(file)
		if len(configPath) == 0 {
			configPath = "./"
		}

		var regexEnv *regexp.Regexp
		var regex *regexp.Regexp

		ext := filepath.Ext(fileName)
		extTrimmed := strings.TrimSuffix(fileName, ext)
		if len(ext) == 0 {
			ext = regexExt
			debugPrintf(darkGrey("\nlooking for '%s%s' in '%s'..."), fileName, regexExt, configPath)
		} else {
			debugPrintf(darkGrey("\nlooking for '%s' in '%s'..."), fileName, configPath)
		}

		format := "^%s%s$"
		if !fileSearchCaseSensitive {
			format = "(?i)(^%s)%s$"
		}
		regexEnv = regexp.MustCompile(fmt.Sprintf(format, fmt.Sprintf("%s.%s", extTrimmed, Env().ID()), ext))
		regex = regexp.MustCompile(fmt.Sprintf(format, extTrimmed, ext))

		// look for the config file in the config path (eg.: tool.yml)
		if matchedFiles := walkConfigPath(configPath, regex); len(matchedFiles) > 0 {
			foundFiles = append(foundFiles, matchedFiles)
		}

		// look for the env config file in the config path (eg.: tool.development.yml)
		if matchedFiles := walkConfigPath(configPath, regexEnv); len(matchedFiles) > 0 {
			foundFiles = append(foundFiles, matchedFiles)
		}
	}

	if len(foundFiles) > 0 {
		debugPrintf("\n%s %s", strings.Join(foundFiles, green(" <- ")), green("="))
	} else {
		debugPrintf("\n")
	}
	return
}

// INTERNAL PARSING ----------------------------------------------------------------------------------------------------

func unmarshalJSON(data []byte, config interface{}, filePath string) (err error) {
	return json.Unmarshal(data, config)
}

func unmarshalTOML(data []byte, config interface{}, filePath string) (err error) {
	_, err = toml.Decode(string(data), config)
	return err
}

func unmarshalYAML(data []byte, config interface{}, filePath string) (err error) {
	return yaml.Unmarshal(data, config)
}

// parseConfigTags will process the struct field tags.
func parseConfigTags(elem interface{}, indent string) error {
	elemValue := reflect.Indirect(reflect.ValueOf(elem))

	switch elemValue.Kind() {

	case reflect.Struct:
		elemType := elemValue.Type()
		verbosePrintf("%sProcessing STRUCT: %s = %+v\n", indent, elemType.Name(), elem)

		for i := 0; i < elemType.NumField(); i++ {

			ft := elemType.Field(i)
			fv := elemValue.Field(i)

			if !fv.CanAddr() || !fv.CanInterface() {
				verbosePrintf("%sCan't addr or interface FIELD: CanAddr: %v, CanInterface: %v. -> %s = '%+v'\n",
					indent, fv.CanAddr(), fv.CanInterface(), ft.Name, fv.Interface())
				continue
			}

			tag := ft.Tag.Get(sftKey)
			tagFields := strings.Split(tag, ",")
			verbosePrintf("\n%sProcessing FIELD: %s %s = %+v, tags: %s\n", indent, ft.Name, ft.Type.String(), fv.Interface(), tag)
			for _, flag := range tagFields {

				kv := strings.Split(flag, "=")

				if kv[0] == sftEnv {
					if len(kv) == 2 {
						if value := os.Getenv(kv[1]); len(value) > 0 {
							debugPrintf("Loading configuration for struct `%v`'s field `%v` from env %v...\n", elemType.Name(), ft.Name, kv[1])
							if err := yaml.Unmarshal([]byte(value), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					}
				}

				if empty := reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()); empty {
					if kv[0] == sftDefault {
						if len(kv) == 2 {
							if err := yaml.Unmarshal([]byte(kv[1]), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					} else if kv[0] == sftRequired {
						return errors.New(ft.Name + " is required")
					}
				}
			}

			switch fv.Kind() {
			case reflect.Ptr, reflect.Struct, reflect.Slice, reflect.Map:
				if err := parseConfigTags(fv.Addr().Interface(), "	"); err != nil {
					return err
				}
			}

			verbosePrintf("%sProcessed  FIELD: %s %s = %+v\n", indent, ft.Name, ft.Type.String(), fv.Interface())
		}

	case reflect.Slice:
		for i := 0; i < elemValue.Len(); i++ {
			if err := parseConfigTags(elemValue.Index(i).Addr().Interface(), "	"); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range elemValue.MapKeys() {
			if err := parseConfigTags(elemValue.MapIndex(key).Interface(), "	"); err != nil {
				return err
			}
		}
	}

	return nil
}

// EXPORTED ------------------------------------------------------------------------------------------------------------

// Unmarshal will unmarshal []byte to interface
// for yaml, toml and json data formats.
//
// Will also parse struct flags.
func Unmarshal(in []byte, out interface{}) (err error) {
	if err = unmarshalJSON(in, out, ""); err == nil {
		return parseConfigTags(out, "")
	}

	if err = unmarshalYAML(in, out, ""); err == nil {
		return parseConfigTags(out, "")
	}

	if err = unmarshalTOML(in, out, ""); err == nil {
		return parseConfigTags(out, "")
	}

	return fmt.Errorf("the provided data is incompatible with an interface of type %T:\n%s",
		out, strings.TrimSuffix(string(in), "\n"))
}

// LoadConfig will unmarshal all the matched
// config files to the config interface.
//
// Build-environment specific files will override generic files.
// The latest files will override the earliest.
//
// Will also parse struct flags.
func LoadConfig(config interface{}, files ...string) (err error) {
	foundFiles := configFilesByEnv(files...)
	if len(foundFiles) == 0 {
		return fmt.Errorf("no config file found for '%s'", strings.Join(files, " | "))
	}

	for _, file := range foundFiles {
		var in []byte
		if in, err = ioutil.ReadFile(file); err != nil {
			return err
		}

		ext := filepath.Ext(file)

		switch {
		case regexp.MustCompile(regexYAML).MatchString(ext):
			err = unmarshalYAML(in, config, file)
		case regexp.MustCompile(regexTOML).MatchString(ext):
			err = unmarshalTOML(in, config, file)
		case regexp.MustCompile(regexJSON).MatchString(ext):
			err = unmarshalJSON(in, config, file)
		default:
			err = fmt.Errorf("unknown data format, can't unmarshal file: '%s'", file)
		}

		if err != nil {
			return err
		}
	}

	defer debugPrintf("\n%s\n", green(dump(config)))
	return parseConfigTags(config, "")
}
