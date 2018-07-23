package sprbox

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// small slant
const banner = `
                __          
  ___ ___  ____/ / ___ __ __
 (_-</ _ \/ __/ _ / _ \\ \ /
/___/ .__/_/ /_.__\___/_\_\  %s
env/_/aware toolbox factory

`

// struct field tags
const (
	sftConfig = "config"
	sftOmit   = "omit"
)

var debug = false

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

	errNoConfigurable = errors.New(`does not implement the 'configurable' interface: SBConfig(string) error`)

	errConfigFileNotFound = errors.New("config file not found")
)

type configurable interface {
	SBConfig(string) error
}

// Load initialize and (eventually) configure the passed struct
// looking for the config files in the passed path.
func Load(box interface{}, path string) error {
	t := reflect.TypeOf(box).Elem()
	v := reflect.ValueOf(box).Elem()

	if t.Kind() != reflect.Struct {
		return errNotAStructPointer
	} else if !v.CanSet() || !v.IsValid() {
		return errInvalidPointer
	}

	debugPrintf("ORIGINAL BOX: %#v\n", box)
	printLoadHeader()
	var err error
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if err = loadBox(path, &ft, ft.Type, fv); err != nil {
			break
		}
	}
	debugPrintf("INITIALIZED BOX: %#v\n", v)
	fmt.Printf("\n")
	return err
}

func loadBox(configPath string, f *reflect.StructField, t reflect.Type, v reflect.Value) error {
	switch t.Kind() {
	case reflect.Ptr:
		debugPrintf("Ptr %#v\n", v)

		configFile, omit := parseTags(f)
		if omit {
			break
		}

		if v.IsNil() {
			debugPrintf("Ptr was nil\n")
			newV := reflect.New(t.Elem())
			v.Set(newV)
		}
		if err := loadConfig(configPath, configFile, f, t, &v); err != nil {
			return err
		}

	case reflect.Struct:
		debugPrintf("Struct %#v\n", v)

		configFile, omit := parseTags(f)
		if omit {
			break
		}

		newV := reflect.New(t)
		if err := loadConfig(configPath, configFile, f, t, &newV); err != nil {
			return err
		}
		v.Set(newV.Elem())

	default:
		break
	}

	return nil
}

// lookupTags returns the config file name and the omit flag.
// The name will be returned also if not specified in tags,
// the field name without extension will be returned in that case,
// loadConfig will look for a file with that prefix and any kind
// of extension, if necessary (no '.' in file name).
func parseTags(f *reflect.StructField) (configFile string, omit bool) {
	configFile = f.Name
	if tag, found := f.Tag.Lookup(sftOmit); found {
		if values := strings.Split(tag, ","); len(values) > 0 {
			for _, value := range values {
				if value == "true" {
					printLoadResult(f, f.Type, errOmit)
					return configFile, true
				}
			}
		}
	}
	if tag, found := f.Tag.Lookup(sftConfig); found {
		if values := strings.Split(tag, ","); len(values) > 0 {
			for _, value := range values {
				configFile = value
			}
		}
	}
	return
}

func loadConfig(configPath string, configFile string, f *reflect.StructField, t reflect.Type, v *reflect.Value) error {
	if _, isConfigurable := v.Interface().(configurable); !isConfigurable {
		printLoadResult(f, t, errNoConfigurable)
		return nil
	}

	if filePath := SearchFileByEnv(configPath, configFile, true); len(filePath) > 0 {
		err := v.Interface().(configurable).SBConfig(filePath)
		printLoadResult(f, t, err)
		return err
	}

	printLoadResult(f, t, errConfigFileNotFound)
	return nil
}

func debugPrintf(format string, args ...interface{}) {
	if debug {
		fmt.Printf(format, args...)
	}
}

// PrintInfo print some useful info about
// the environment and git on init.
func PrintInfo(hideBanner bool) {
	if !hideBanner {
		version := ""
		sprboxRepo := NewRepository(filepath.Join(os.Getenv("GOPATH"), "/src/github.com/oblq/sprbox"))
		if sprboxRepo.Error == nil {
			version = "v" + sprboxRepo.Tag + "(" + sprboxRepo.Build + ")"
		} else {
			println(sprboxRepo.Error.Error())
		}
		fmt.Printf(darkGrey(banner), version)
	}

	Env().PrintInfo()
	VCS.PrintInfo()
}
