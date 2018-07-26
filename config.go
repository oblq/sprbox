package sprbox

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"os"
	"reflect"
	"strings"

	"fmt"

	"encoding/json"

	"bytes"

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
		debugPrintf("walkConfigPath error: %s", err.Error())
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
func configFilesByEnv(filePath string) (files []string) {
	configPath, file := filepath.Split(filePath)
	if len(configPath) == 0 {
		configPath = "./"
	}

	var regexEnv *regexp.Regexp
	var regex *regexp.Regexp

	ext := filepath.Ext(file)
	extTrimmed := strings.TrimSuffix(file, ext)
	if len(ext) == 0 {
		ext = regexExt
		debugPrintf(darkGrey("\nLooking for '%s.*' in '%s':"), file, configPath)
	} else {
		debugPrintf(darkGrey("\nLooking for '%s' in '%s':"), file, configPath)
	}

	format := "^%s%s$"
	if !FileSearchCaseSensitive {
		format = "(?i)(^%s)%s$"
	}
	regexEnv = regexp.MustCompile(fmt.Sprintf(format, fmt.Sprintf("%s.%s", extTrimmed, Env().String()), ext))
	regex = regexp.MustCompile(fmt.Sprintf(format, extTrimmed, ext))

	// look for the config file in the config path (eg.: tool.yml)
	if matchedFiles := walkConfigPath(configPath, regex); len(matchedFiles) > 0 {
		files = append(files, matchedFiles)
	}

	// look for the env config file in the config path (eg.: tool.development.yml)
	if matchedFiles := walkConfigPath(configPath, regexEnv); len(matchedFiles) > 0 {
		files = append(files, matchedFiles)
	}

	debugPrintf("\n%s", strings.Join(files, green(" <- ")))
	return
}

// mergedConfigs returns all the matched config files merged in the right order.
// (eg.: conf.<environment>.yml -> conf.yml)
func mergedConfigs(filePath string) (data []byte, err error) {
	files := configFilesByEnv(filePath)
	if len(files) == 0 {
		return nil, fmt.Errorf("config file(s) not found for '%s'", filePath)
	}

	var merged map[string]interface{}
	for _, file := range files {
		if err := unmarshal(file, nil, &merged); err != nil {
			return nil, err
		}
	}

	debugPrintf(green(" = ")+"%+v\n", green(prettyPrinted(merged)))
	ext := filepath.Ext(files[0])

	switch {
	case regexp.MustCompile(regexYAML).MatchString(ext):
		return yaml.Marshal(merged)
	case regexp.MustCompile(regexTOML).MatchString(ext):
		var buffer bytes.Buffer
		err := toml.NewEncoder(&buffer).Encode(merged)
		return buffer.Bytes(), err
	case regexp.MustCompile(regexJSON).MatchString(ext):
		return json.Marshal(merged)
	default:
		return nil, errors.New("unknown config file format")
	}
}

// INTERNAL PARSING ----------------------------------------------------------------------------------------------------

func unmarshalJSON(data []byte, config interface{}, filePath string) (err error) {
	err = json.Unmarshal(data, config)
	return
}

func unmarshalTOML(data []byte, config interface{}, filePath string) (err error) {
	_, err = toml.Decode(string(data), config)
	return err
}

func unmarshalYAML(data []byte, config interface{}, filePath string) (err error) {
	err = yaml.Unmarshal(data, config)
	return
}

// unmarshal will unmarshall the file or the file bytes to the 'out' interface.
// 'filePath' is not mandatory, if used must include file extension.
func unmarshal(filePath string, in []byte, out interface{}) (err error) {
	if len(filePath) > 0 {
		if in, err = ioutil.ReadFile(filePath); err != nil {
			return err
		}
	}

	ext := filepath.Ext(filePath)

	if len(ext) > 0 {
		switch {
		case regexp.MustCompile(regexYAML).MatchString(ext):
			return unmarshalYAML(in, out, filePath)
		case regexp.MustCompile(regexTOML).MatchString(ext):
			return unmarshalTOML(in, out, filePath)
		case regexp.MustCompile(regexJSON).MatchString(ext):
			return unmarshalJSON(in, out, filePath)
		default:
			return fmt.Errorf("unknown data format, can't unmarshal file: '%s'", filePath)
		}
	} else {
		if err = unmarshalJSON(in, out, filePath); err == nil {
			return nil
		}

		if err = unmarshalYAML(in, out, filePath); err == nil {
			return nil
		}

		if err = unmarshalTOML(in, out, filePath); err == nil {
			return nil
		}

		return fmt.Errorf("unknown data format, can't unmarshal file content: '%s'",
			strings.TrimSuffix(string(in), "\n"))
	}
}

// parseConfigTags will process the struct field tags.
func parseConfigTags(config interface{}) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct: " + configValue.Kind().String())
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {

		ft := configType.Field(i)
		fv := configValue.Field(i)

		if !fv.CanAddr() || !fv.CanInterface() {
			continue
		}

		tag := ft.Tag.Get(sftKey)
		fields := strings.Split(tag, ",")
		for _, flag := range fields {

			kv := strings.Split(flag, "=")
			if len(kv) < 1 {
				continue
			}

			if kv[0] == sftEnv {
				if len(kv) == 2 {
					if value := os.Getenv(kv[1]); len(value) > 0 {
						debugPrintf("Loading configuration for struct `%v`'s field `%v` from env %v...\n", configType.Name(), ft.Name, kv[1])
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
						continue
					}
				} else if kv[0] == sftRequired {
					return errors.New(ft.Name + " is required")
				}
			}
		}

		// recursive check
		{
			for fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}

			if fv.Kind() == reflect.Struct {
				if err := parseConfigTags(fv.Addr().Interface()); err != nil {
					return err
				}
			}

			if fv.Kind() == reflect.Slice {
				for i := 0; i < fv.Len(); i++ {
					if reflect.Indirect(fv.Index(i)).Kind() == reflect.Struct {
						if err := parseConfigTags(fv.Index(i).Addr().Interface()); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

// EXPORTED ------------------------------------------------------------------------------------------------------------

// Unmarshal will unmarshal []bytes to interface
// for yaml, toml and json data formats.
// Useful in the 'configurable' interface in which
// the received bytes are the result of multiple
// merged config files.
// Will also parse struct flags.
func Unmarshal(in []byte, out interface{}) (err error) {
	if err = unmarshal("", in, out); err == nil {
		err = parseConfigTags(out)
	}
	return
}

// LoadConfig will unmarshal the provided config file
// eventually overriding it with an environment specific one,
// if present, to the provided struct pointer.
// Will also parse struct flags.
func LoadConfig(config interface{}, filePath string) (err error) {
	var in []byte
	if in, err = mergedConfigs(filePath); err != nil {
		return
	}

	if err = unmarshal("", in, config); err == nil {
		err = parseConfigTags(config)
	}

	defer fmt.Print("\n")
	defer debugPrintf("%s%s\n", "Loaded config: ", green(prettyPrinted(config)))

	return
}
