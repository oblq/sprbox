package sprbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Struct field flags.
const (
	sftOmit = "omit"
)

// Errors.
var (
	errInvalidPointer             = errors.New("<box> parameter should be a struct pointer")
	errOmit                       = errors.New("omitted")
	errNoConfigurable             = errors.New(`does not implement the 'configurable' interface: func SpareConfig([]string) error`)
	errNoConfigurableInCollection = errors.New(`does not implement the 'configurableInCollection' interface: func SpareConfigBytes([]byte) error`)
	errConfigFileNotFound         = errors.New("config file not found")
)

type configurable interface {
	SpareConfig([]string) error
}

type configurableInCollection interface {
	SpareConfigBytes([]byte) error
}

// LoadToolBox initialize and (eventually) configure the provided struct pointer
// looking for the config files in the provided configPath.
func LoadToolBox(toolBox interface{}, configPath string) (err error) {
	t := reflect.TypeOf(toolBox).Elem()
	v := reflect.ValueOf(toolBox).Elem()

	if t.Kind() != reflect.Struct {
		return errInvalidPointer
	} else if !v.CanSet() || !v.IsValid() {
		return errInvalidPointer // nil pointer
	}

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)
		if err = loadField(configPath, &sf, fv, ""); err != nil {
			break
		}
	}
	debugPrintf("\nLoaded toolbox: \n%s\n", green(dump(toolBox)))
	fmt.Print("\n")
	return
}

func loadField(configPath string, sf *reflect.StructField, fv reflect.Value, indent string) error {
	switch fv.Kind() {
	case reflect.Ptr:
		fv.Set(reflect.New(fv.Type().Elem()))
		return loadField(configPath, sf, fv.Elem(), "")

	case reflect.Struct:
		configFiles := []string{sf.Name}
		if omit := parseTags(&configFiles, sf); omit {
			break
		}

		newV := reflect.New(fv.Type())

		if _, isConfigurable := newV.Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, &newV, indent); err != nil {
				return err
			}

			fv.Set(newV.Elem())
		} else {
			printLoadResult(sf.Name, sf.Type, errNoConfigurable)
			fv.Set(newV.Elem())
			// recursive
			for i := 0; i < fv.NumField(); i++ {
				ssf := fv.Type().Field(i)
				sfv := fv.Field(i)
				if err := loadField(configPath, &ssf, sfv, "  "); err != nil {
					return err
				}
			}
			return nil
		}

	case reflect.Slice:
		configFiles := []string{sf.Name}
		if omit := parseTags(&configFiles, sf); omit {
			break
		}

		newV := reflect.New(fv.Type())

		if _, isConfigurable := newV.Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, &newV, indent); err != nil {
				return err
			}
			fv.Set(newV.Elem())
		} else {
			printLoadResult(sf.Name, sf.Type, errNoConfigurable)

			for i, file := range configFiles {
				configFiles[i] = filepath.Join(configPath, file)
			}

			var config []interface{}
			if err := LoadConfig(&config, configFiles...); err != nil {
				printLoadResult("  "+sf.Name, sf.Type.Elem(), err)
				return err
			}

			newV := reflect.New(fv.Type()).Elem()

			for i := 0; i < len(config); i++ {
				elemType := fv.Type().Elem()
				var elem reflect.Value
				sfName := fmt.Sprintf("%s[%d]", sf.Name, i)

				switch elemType.Kind() {
				case reflect.Ptr:
					elem = reflect.New(elemType.Elem())
					if err := configureElem(elem, config[i], sfName); err != nil {
						return err
					}
					printLoadResult("  "+sfName, elem.Type(), nil)
					newV.Set(reflect.Append(newV, elem))

				case reflect.Struct:
					elem = reflect.New(elemType)
					if err := configureElem(elem, config[i], sfName); err != nil {
						return err
					}
					printLoadResult("  "+sfName, elem.Elem().Type(), nil)
					newV.Set(reflect.Append(newV, elem.Elem()))
				}
			}

			fv.Set(newV)
			return nil
		}

	case reflect.Map:
		configFiles := []string{sf.Name}
		if omit := parseTags(&configFiles, sf); omit {
			break
		}

		newV := reflect.New(fv.Type())

		if _, isConfigurable := newV.Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, &newV, indent); err != nil {
				return err
			}
			fv.Set(newV.Elem())
		} else {
			printLoadResult(sf.Name, sf.Type, errNoConfigurable)

			for i, file := range configFiles {
				configFiles[i] = filepath.Join(configPath, file)
			}

			var config map[string]interface{}
			if err := LoadConfig(&config, configFiles...); err != nil {
				printLoadResult("  "+sf.Name, fv.Type(), err)
				return err
			}

			newV := reflect.MakeMapWithSize(fv.Type(), len(config))

			for key, conf := range config {
				kv := reflect.ValueOf(key)

				elemType := fv.Type().Elem()
				var elem reflect.Value
				sfName := fmt.Sprintf("%s[%s]", sf.Name, key)

				switch elemType.Kind() {
				case reflect.Ptr:
					elem = reflect.New(elemType.Elem())
					if err := configureElem(elem, conf, sfName); err != nil {
						return err
					}
					printLoadResult("  "+sfName, elem.Type(), nil)
					newV.SetMapIndex(kv, elem)

				case reflect.Struct:
					elem = reflect.New(elemType)
					if err := configureElem(elem, conf, sfName); err != nil {
						return err
					}
					printLoadResult("  "+sfName, elem.Elem().Type(), nil)
					newV.SetMapIndex(kv, elem.Elem())
				}
			}

			fv.Set(newV)
			return nil
		}

	default:
		break
	}

	return nil
}

// parseTags returns the config file name and the omit flag.
// The name will be returned also if not specified in tags,
// the field name without extension will be returned in that case,
// loadConfig will look for a file with that prefix and any kind
// of extension, if necessary (no '.' in file name).
func parseTags(configFiles *[]string, f *reflect.StructField) (omit bool) {
	tag, found := f.Tag.Lookup(sftKey)
	if !found {
		return
	}

	if regexp.MustCompile(sftOmit).MatchString(tag) {
		//printLoadResult(f.Name, f.Type, errOmit)
		return true
	}

	tagFields := strings.Split(tag, ",")
	for _, flag := range tagFields {
		if flag != sftOmit {
			files := strings.Split(flag, "|")
			*configFiles = append(*configFiles, files...)
		}
	}

	return
}

// configure will call the 'configurable' interface on the passed field struct pointer.
func configure(configPath string, configFiles []string, f *reflect.StructField, v *reflect.Value, indent string) error {
	for i, file := range configFiles {
		configFiles[i] = filepath.Join(configPath, file)
	}

	if err := v.Interface().(configurable).SpareConfig(configFiles); err != nil {
		printLoadResult(indent+f.Name, f.Type, err)
		return err
	}

	printLoadResult(indent+f.Name, f.Type, nil)
	return nil
}

// configureElem will call the 'configurableInCollection' interface on the passed struct pointer.
func configureElem(elem reflect.Value, config interface{}, sfName string) (err error) {
	if _, isConfigurable := elem.Interface().(configurableInCollection); !isConfigurable {
		printLoadResult(sfName, elem.Type(), errNoConfigurableInCollection)
		return nil
	}

	var bytes []byte
	if bytes, err = json.Marshal(config); err != nil {
		if bytes, err = yaml.Marshal(config); err != nil {
			printLoadResult(sfName, elem.Type(), err)
			return err
		}
	}

	if err = elem.Interface().(configurableInCollection).SpareConfigBytes(bytes); err != nil {
		printLoadResult(sfName, elem.Type(), err)
		return err
	}

	return nil
}

func printLoadResult(objNameType string, t reflect.Type, err error) {
	if len(objNameType) == 0 {
		objNameType = t.Name()
	}
	objType := t.String()
	if len(objType) > 50 {
		objType = t.Kind().String()
	}
	objNameType = fmt.Sprintf("%v (%v)", blue(objNameType), objType)
	objNameType = fmt.Sprintf("%-50v", objNameType)
	if err != nil {
		if err == errOmit {
			fmt.Printf("%s %s\n", objNameType, err.Error())
		} else if err == errNoConfigurable ||
			err == errNoConfigurableInCollection ||
			err == errConfigFileNotFound {
			fmt.Printf("%s %s\n", objNameType, yellow("-> "+err.Error()))
		} else {
			fmt.Printf("%s %s\n", objNameType, red("-> "+err.Error()))
		}
	} else {
		fmt.Printf("%s %s\n", objNameType, green("<- config loaded"))
	}
}
