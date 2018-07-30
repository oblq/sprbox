package sprbox

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// struct field flags
const (
	sftOmit = "omit"
)

var (
	errNotAStructPointer = errors.New("<box> should be a pointer to a struct")

	errInvalidPointer = errors.New(`
	
	invalid <box> pointer, don't do:
		var MyAppToolBox *Box
		InitAndConfig(MyAppToolBox, "path/to/config")
	
	Init the pointer before:
		var MyAppToolBox = &Box{}
		InitAndConfig(MyAppToolBox, "path/to/config")
	
	...or pass a new pointer:
		var MyAppToolBox Box
		InitAndConfig(&MyAppToolBox, "path/to/config")`)

	errOmit = errors.New("omitted")

	errNoConfigurable = errors.New(`does not implement the 'configurable' interface: func SBConfig([]byte) error`)

	errConfigFileNotFound = errors.New("config file not found")
)

type configurable interface {
	SBConfig([]byte) error
}

// LoadToolBox initialize and (eventually) configure the provided struct pointer
// looking for the config files in the provided configPath.
func LoadToolBox(toolBox interface{}, configPath string) error {
	t := reflect.TypeOf(toolBox).Elem()
	v := reflect.ValueOf(toolBox).Elem()

	if t.Kind() != reflect.Struct {
		return errNotAStructPointer
	} else if !v.CanSet() || !v.IsValid() {
		return errInvalidPointer
	}

	//debugPrintf("ORIGINAL BOX: %#v\n", toolBox)
	//printLoadHeader()
	var err error
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if err = loadField(configPath, &ft, ft.Type, fv); err != nil {
			break
		}
	}
	debugPrintf("\nLoaded toolbox: %s\n", green(dump(toolBox)))
	fmt.Print("\n")
	return err
}

func loadField(configPath string, f *reflect.StructField, t reflect.Type, v reflect.Value) error {
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			newV := reflect.New(t.Elem())
			v.Set(newV)
		}

		configFiles := []string{f.Name}
		if omit := parseTags(&configFiles, f); omit {
			break
		}

		if err := configure(configPath, configFiles, f, t, &v); err != nil {
			return err
		}

	case reflect.Struct:
		configFiles := []string{f.Name}
		if omit := parseTags(&configFiles, f); omit {
			break
		}

		newV := reflect.New(t)
		if err := configure(configPath, configFiles, f, t, &newV); err != nil {
			return err
		}
		v.Set(newV.Elem())

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
		printLoadResult(f, f.Type, errOmit)
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

// configure will call the 'configurable' interface
// on the passed field struct.
func configure(configPath string, configFiles []string, f *reflect.StructField, t reflect.Type, v *reflect.Value) error {
	if _, isConfigurable := v.Interface().(configurable); !isConfigurable {
		printLoadResult(f, t, errNoConfigurable)
		return nil
	}

	// add configPath to any file name
	for i, file := range configFiles {
		configFiles[i] = filepath.Join(configPath, file)
	}

	bytes, err := mergedConfigs(configFiles)
	if err != nil {
		printLoadResult(f, t, err)
		return nil
	}

	if err := v.Interface().(configurable).SBConfig(bytes); err != nil {
		printLoadResult(f, t, err)
		return err
	}

	printLoadResult(f, t, nil)
	return nil
}

func printLoadResult(f *reflect.StructField, t reflect.Type, err error) {
	objNameType := t.Name()
	if f != nil {
		objNameType = f.Name
	}
	objNameType = fmt.Sprintf("%v (%v)", blue(objNameType), t.String())
	objNameType = fmt.Sprintf("%-50v", objNameType)
	if err != nil {
		if err == errOmit {
			fmt.Printf("%s %s\n", objNameType, err.Error())
		} else if err == errNoConfigurable || err == errConfigFileNotFound {
			fmt.Printf("%s %s\n", objNameType, yellow("-> "+err.Error()))
		} else {
			fmt.Printf("%s %s\n", objNameType, red("-> "+err.Error()))
		}
	} else {
		fmt.Printf("%s %s\n", objNameType, green("<- config loaded"))
	}
}
