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

const (
	structFTag      = "sprbox"
	structFTagVOmit = "omit"
)

var debug = false

var (
	errNotAStructPointer = errors.New("<box> must be a pointer to a struct")

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

	errNoBoxable = errors.New(`does not implement the 'boxable' interface: Go2Box(string) error`)
)

type configurable interface {
	SBConfig(string) error
}

// InitAndConfig initialize and configure the passed struct
// looking for the config files in the passed path.
func InitAndConfig(box interface{}, path string) error {
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
	if _, isConfigurable := reflect.ValueOf(box).Interface().(configurable); isConfigurable {
		err = initBox(SubPathByEnv(path), nil, t, v)
	} else {
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			fv := v.Field(i)
			err = initBox(SubPathByEnv(path), &ft, ft.Type, fv)
		}
	}
	debugPrintf("INITIALIZED BOX: %#v\n", v)
	fmt.Printf("\n")
	return err
}

func initBox(configPath string, f *reflect.StructField, t reflect.Type, v reflect.Value) error {
	switch t.Kind() {
	case reflect.Ptr:
		debugPrintf("Ptr %#v\n", v)

		configFile, omit := lookupTags(f)
		if omit {
			break
		}

		if v.IsNil() {
			debugPrintf("Ptr was nil\n")
			newV := reflect.New(t.Elem())
			v.Set(newV)
		}
		config := filepath.Join(configPath, configFile)
		if err := loadConfig(config, f, t, &v); err != nil {
			return err
		}

	case reflect.Struct:
		debugPrintf("Struct %#v\n", v)

		configFile, omit := lookupTags(f)
		if omit {
			break
		}

		newV := reflect.New(t)
		config := filepath.Join(configPath, configFile)
		if err := loadConfig(config, f, t, &newV); err != nil {
			return err
		}
		v.Set(newV.Elem())

	default:
		break
	}

	return nil
}

func lookupTags(f *reflect.StructField) (configFile string, omit bool) {
	if f == nil {
		return
	}
	configFile = f.Name + ".yml"
	if tag, found := f.Tag.Lookup(structFTag); found {
		if values := strings.Split(tag, ","); len(values) > 0 {
			for _, value := range values {
				if value == structFTagVOmit {
					printLoadResult(f, f.Type, errOmit)
					return configFile, true
				}
				configFile = value
			}
		}
	}
	return
}

func loadConfig(configPath string, f *reflect.StructField, t reflect.Type, v *reflect.Value) error {
	if _, isConfigurable := v.Interface().(configurable); !isConfigurable {
		printLoadResult(f, t, errNoBoxable)
		return nil
	}
	err := v.Interface().(configurable).SBConfig(configPath)
	printLoadResult(f, t, err)
	return err
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
		gomsVersion := ""
		sprboxRepo := NewRepository(filepath.Join(os.Getenv("GOPATH"), "/src/github.com/oblq/sprbox"))
		if sprboxRepo.Error == nil {
			gomsVersion = "v" + sprboxRepo.Tag + "(" + sprboxRepo.Build + ")"
		} else {
			println(sprboxRepo.Error.Error())
		}
		fmt.Printf(darkGrey(banner), gomsVersion)
	}

	Env().PrintInfo()
	VCS.PrintInfo()
}
