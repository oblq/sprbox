package sprbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

// struct field flags
const (
	// sffEnv value can be in json format, it will override also the default value
	sffEnv = "env"

	// set the default value
	sffDefault = "default"

	// return error if missing value
	sffRequired = "required"
)

// files type regexp
const (
	regexYAML = `(?i)(.y(|a)ml)`
	regexTOML = `(?i)(.toml)`
	regexJSON = `(?i)(.json)`
)

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
			verbosePrintf("\n%sProcessing FIELD: %s %s = %+v, tags: %s\n",
				indent, ft.Name, ft.Type.String(), fv.Interface(), tag)
			for _, flag := range tagFields {

				kv := strings.Split(flag, "=")

				if kv[0] == sffEnv {
					if len(kv) == 2 {
						if value := os.Getenv(kv[1]); len(value) > 0 {
							debugPrintf("Loading configuration for struct `%v`'s field `%v` from env %v...\n",
								elemType.Name(), ft.Name, kv[1])
							if err := yaml.Unmarshal([]byte(value), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					}
				}

				if empty := reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()); empty {
					if kv[0] == sffDefault {
						if len(kv) == 2 {
							if err := yaml.Unmarshal([]byte(kv[1]), fv.Addr().Interface()); err != nil {
								return err
							}
						}
					} else if kv[0] == sffRequired {
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

// parseTemplateBytes parse all text/template placeholders
// (eg.: {{.Key}}) in config files.
func parseTemplateBytes(file []byte, config interface{}) error {
	var buf bytes.Buffer
	var tpl *template.Template
	var err error

	if tpl, err = template.New("tpl").Parse(string(file)); err != nil {
		return err
	}

	if err = tpl.Execute(&buf, config); err != nil {
		return err
	}

	switch {
	case unmarshalJSON(buf.Bytes(), config, "") == nil:
		return nil
	case unmarshalYAML(buf.Bytes(), config, "") == nil:
		return nil
	case unmarshalTOML(buf.Bytes(), config, "") == nil:
		return nil
	default:
		return fmt.Errorf("the provided data is incompatible with an interface of type %T:\n%s",
			config, strings.TrimSuffix(string(file), "\n"))
	}
}

// parseTemplateFile parse all text/template placeholders
// (eg.: {{.Key}}) in config files.
func parseTemplateFile(file string, config interface{}) error {
	var buf bytes.Buffer
	var tpl *template.Template
	var err error

	if tpl, err = template.ParseFiles(file); err != nil {
		return err
	}
	if err = tpl.Execute(&buf, config); err != nil {
		return err
	}

	ext := filepath.Ext(file)

	switch {
	case regexp.MustCompile(regexYAML).MatchString(ext):
		return unmarshalYAML(buf.Bytes(), config, file)
	case regexp.MustCompile(regexTOML).MatchString(ext):
		return unmarshalTOML(buf.Bytes(), config, file)
	case regexp.MustCompile(regexJSON).MatchString(ext):
		return unmarshalJSON(buf.Bytes(), config, file)
	default:
		return fmt.Errorf("unknown data format, can't unmarshal file: '%s'", file)
	}
}

// Unmarshal will unmarshal []byte to interface
// for yaml, toml and json data formats.
//
// Will also parse struct flags.
func Unmarshal(data []byte, config interface{}) (err error) {
	switch {
	case unmarshalJSON(data, config, "") == nil:
		break
	case unmarshalYAML(data, config, "") == nil:
		break
	case unmarshalTOML(data, config, "") == nil:
		break
	default:
		return fmt.Errorf("the provided data is incompatible with an interface of type %T:\n%s",
			config, strings.TrimSuffix(string(data), "\n"))
	}

	//debugPrintf("elem: %s\n%+v\n", string(data), config)

	if err = parseTemplateBytes(data, config); err != nil {
		return err
	}
	return parseConfigTags(config, "")
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

		if err = parseTemplateFile(file, config); err != nil {
			return err
		}
	}

	//configB, err := yaml.Marshal(config)
	//if err != nil {
	//	return err
	//}
	//if err = parseTemplate(configB, config); err != nil {
	//	return err
	//}

	defer debugPrintf("%s\n", green(dump(config)))
	return parseConfigTags(config, "")
}
