package sprbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// Struct field flags.
const (
	sftSkip = "-"
)

// Errors.
var (
	errInvalidPointer              = errors.New("<box> parameter should be a struct pointer")
	errNotConfigurable             = errors.New("does not implement the 'configurable' interface: `func SpareConfig([]string) error`")
	errNotConfigurableInCollection = errors.New("does not implement the 'configurable' interface nor its elements implements the 'configurableInCollection' one: `func SpareConfigBytes([]byte) error`")
)

type configurable interface {
	SpareConfig([]string) error
}

type configurableInCollection interface {
	SpareConfigBytes([]byte) error
}

// If PkgPath is set, the field is not exported
//	exported := field.PkgPath == ""

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
		if err = loadField(configPath, &sf, fv, 0); err != nil {
			break
		}
	}
	debugPrintf("\nLoaded toolbox: \n%s\n", green(dump(toolBox)))
	fmt.Print("\n")
	return
}

// level is the parent grade to the initially passed fv
func loadField(configPath string, sf *reflect.StructField, fv reflect.Value, level int) error {
	switch fv.Kind() {
	case reflect.Ptr:
		if tag, found := sf.Tag.Lookup(sftKey); found && tag == sftSkip {
			return nil
		}

		if !fv.CanSet() || sf.Anonymous {
			return nil
		}
		// skip already initialized pointers (as can be a '*Config'
		// field configured in 'configurable' interface call).
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return loadField(configPath, sf, fv.Elem(), level)

	case reflect.Struct:
		// !reflect.Zero(fv.Type()) is an already configured field, so sprbox will skip it.
		if !fv.CanSet() || sf.Anonymous ||
			!reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()) {
			return nil
		}

		configFiles := []string{sf.Name}
		if skip := parseTags(&configFiles, sf); skip {
			return nil
		}

		fv.Set(reflect.New(fv.Type()).Elem())

		if _, isConfigurable := fv.Addr().Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, fv.Addr(), level); err != nil {
				return err
			}
		} else {
			printLoadResult(sf.Name, sf.Type, errNotConfigurable, level)
		}

		level += 1
		for i := 0; i < fv.NumField(); i++ {
			ssf := fv.Type().Field(i)
			sfv := fv.Field(i)
			verbosePrintf("%ssub-field: %s\n", strings.Repeat(" -> ", level), ssf.Name)
			//subPath := filepath.Join(configPath, sf.Name)
			if err := loadField(configPath, &ssf, sfv, level); err != nil {
				return err
			}
		}

		return nil

	case reflect.Slice:
		// !reflect.Zero(fv.Type()) is an already configured field, so sprbox will skip it.
		if !fv.CanSet() || sf.Anonymous ||
			!reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()) {
			return nil
		}

		configFiles := []string{sf.Name}
		if skip := parseTags(&configFiles, sf); skip {
			return nil
		}

		fv.Set(reflect.New(fv.Type()).Elem())

		if _, isConfigurable := fv.Addr().Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, fv.Addr(), level); err != nil {
				return err
			}
		} else {
			// skip slices of non-configurableInCollection objects
			cicType := reflect.TypeOf((*configurableInCollection)(nil)).Elem()
			if !fv.Type().Elem().Implements(cicType) &&
				!reflect.PtrTo(fv.Type().Elem()).Implements(cicType) {
				printLoadResult(sf.Name, sf.Type, errNotConfigurableInCollection, level)
				return nil
			}

			printLoadResult(sf.Name, sf.Type, errNotConfigurable, level)

			level += 1

			for i, file := range configFiles {
				configFiles[i] = filepath.Join(configPath, file)
			}

			var config []interface{}
			if err := LoadConfig(&config, configFiles...); err != nil {
				printLoadResult(sf.Name, sf.Type.Elem(), err, level)
				return err
			}

			for i := 0; i < len(config); i++ {
				elemType := fv.Type().Elem()
				var elem reflect.Value
				sfName := fmt.Sprintf("%s[%d]", sf.Name, i)

				switch elemType.Kind() {
				case reflect.Ptr:
					elem = reflect.New(elemType.Elem())
					if err := configureElem(elem, config[i], sfName, level); err != nil {
						return err
					}
					printLoadResult(sfName, elem.Type(), nil, level)
					fv.Set(reflect.Append(fv, elem))

				case reflect.Struct:
					elem = reflect.New(elemType)
					if err := configureElem(elem, config[i], sfName, level); err != nil {
						return err
					}
					printLoadResult(sfName, elem.Elem().Type(), nil, level)
					fv.Set(reflect.Append(fv, elem.Elem()))
				}
			}
		}

		return nil

	case reflect.Map:
		// !reflect.Zero(fv.Type()) is an already configured field, so sprbox will skip it.
		if !fv.CanSet() || sf.Anonymous ||
			!reflect.DeepEqual(fv.Interface(), reflect.Zero(fv.Type()).Interface()) {
			return nil
		}

		configFiles := []string{sf.Name}
		if skip := parseTags(&configFiles, sf); skip {
			return nil
		}

		fv.Set(reflect.New(fv.Type()).Elem())

		if _, isConfigurable := fv.Addr().Interface().(configurable); isConfigurable {
			if err := configure(configPath, configFiles, sf, fv.Addr(), level); err != nil {
				return err
			}
		} else {
			// skip maps of non-configurableInCollection objects
			cicType := reflect.TypeOf((*configurableInCollection)(nil)).Elem()
			if !fv.Type().Elem().Implements(cicType) &&
				!reflect.PtrTo(fv.Type().Elem()).Implements(cicType) {
				printLoadResult(sf.Name, sf.Type, errNotConfigurableInCollection, level)
				return nil
			}

			printLoadResult(sf.Name, sf.Type, errNotConfigurable, level)

			level += 1

			for i, file := range configFiles {
				configFiles[i] = filepath.Join(configPath, file)
			}

			var config map[string]interface{}
			if err := LoadConfig(&config, configFiles...); err != nil {
				printLoadResult(sf.Name, fv.Type(), err, level)
				return err
			}

			fv.Set(reflect.MakeMapWithSize(fv.Type(), len(config)))

			for key, conf := range config {
				kv := reflect.ValueOf(key)

				elemType := fv.Type().Elem()
				var elem reflect.Value
				sfName := fmt.Sprintf("%s[%s]", sf.Name, key)

				switch elemType.Kind() {
				case reflect.Ptr:
					elem = reflect.New(elemType.Elem())
					if err := configureElem(elem, conf, sfName, level); err != nil {
						return err
					}
					printLoadResult(sfName, elem.Type(), nil, level)
					fv.SetMapIndex(kv, elem)

				case reflect.Struct:
					elem = reflect.New(elemType)
					if err := configureElem(elem, conf, sfName, level); err != nil {
						return err
					}
					printLoadResult(sfName, elem.Elem().Type(), nil, level)
					fv.SetMapIndex(kv, elem.Elem())
				}
			}
		}

		return nil

	default:
		return nil
	}
}

// parseTags returns the config file name and the skip flag.
// The name will be returned also if not specified in tags,
// the field name without extension will be returned in that case,
// loadConfig will look for a file with that prefix and any kind
// of extension, if necessary (no '.' in file name).
func parseTags(configFiles *[]string, f *reflect.StructField) (skip bool) {
	tag, found := f.Tag.Lookup(sftKey)
	if !found {
		return
	}

	if tag == sftSkip {
		//printLoadResult(f.Name, f.Type, errOmit)
		return true
	}

	tagFields := strings.Split(tag, ",")
	for _, flag := range tagFields {
		files := strings.Split(flag, "|")
		*configFiles = append(*configFiles, files...)
	}

	return
}

// configure will call the 'configurable' interface on the passed field struct pointer.
func configure(configPath string, configFiles []string, f *reflect.StructField, v reflect.Value, level int) error {
	for i, file := range configFiles {
		configFiles[i] = filepath.Join(configPath, file)
	}

	if err := v.Interface().(configurable).SpareConfig(configFiles); err != nil {
		printLoadResult(f.Name, f.Type, err, level)
		return err
	}

	printLoadResult(f.Name, f.Type, nil, level)
	return nil
}

// configureElem will call the 'configurableInCollection' interface on the passed struct pointer.
func configureElem(elem reflect.Value, config interface{}, sfName string, level int) (err error) {
	var bytes []byte
	if bytes, err = json.Marshal(config); err != nil {
		if bytes, err = yaml.Marshal(config); err != nil {
			printLoadResult(sfName, elem.Type(), err, level)
			return err
		}
	}

	if err = elem.Interface().(configurableInCollection).SpareConfigBytes(bytes); err != nil {
		printLoadResult(sfName, elem.Type(), err, level)
		return err
	}

	return nil
}

func printLoadResult(objNameType string, t reflect.Type, err error, level int) {
	if len(objNameType) == 0 {
		objNameType = t.Name()
	}
	objNameType = strings.Repeat(" -> ", level) + objNameType

	objType := t.String()
	if len(objType)+len(objNameType)+1 >= 60 {
		objType = t.Kind().String()
	}
	objNameType = fmt.Sprintf("%v (%v)", blue(objNameType), objType)
	objNameType = fmt.Sprintf("%-60v", objNameType)
	if err != nil {
		if err == errNotConfigurable || err == errNotConfigurableInCollection {
			if level > 0 { //&& !debug {
				return
			}
			fmt.Printf("%s %s\n", objNameType, yellow("-> "+err.Error()))
		} else {
			fmt.Printf("%s %s\n", objNameType, red("-> "+err.Error()))
		}
	} else {
		fmt.Printf("%s %s\n", objNameType, green("<- config loaded"))
	}
}
