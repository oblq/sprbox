package sprbox

import (
	"fmt"
	"reflect"
)

// ColoredLog enable or disable colors in console.
var ColoredLog = true

type color string

// Color ANSI codes
const (
	defaultFG color = "39"

	redCol      color = "31"
	greenCol    color = "32"
	yellowCol   color = "33"
	blueCol     color = "34"
	magentaCol  color = "35"
	darkGreyCol color = "90"

	esc   = "\033["
	clear = "\033[0m"
)

type painter func(interface{}) string

// Black return the argument as a color escaped string
func def(arg interface{}) string {
	return colorize(fmt.Sprint(arg), defaultFG)
}

// Red return the argument as a color escaped string
func red(arg interface{}) string {
	return colorize(fmt.Sprint(arg), redCol)
}

// Green return the argument as a color escaped string
func green(arg interface{}) string {
	return colorize(fmt.Sprint(arg), greenCol)
}

// Yellow return the argument as a color escaped string
func yellow(arg interface{}) string {
	return colorize(fmt.Sprint(arg), yellowCol)
}

// Blue return the argument as a color escaped string
func blue(arg interface{}) string {
	return colorize(fmt.Sprint(arg), blueCol)
}

// Magenta return the argument as a color escaped string
func magenta(arg interface{}) string {
	return colorize(fmt.Sprint(arg), magentaCol)
}

// DarkGrey return the argument as a color escaped string
func darkGrey(arg interface{}) string {
	return colorize(fmt.Sprint(arg), darkGreyCol)
}

// colored return the ANSI colored formatted string.
func colorize(arg string, color color) string {
	coloredFormat := "%v"
	if len(color) > 0 && ColoredLog {
		coloredFormat = esc + "%vm%v" + clear
		return fmt.Sprintf(coloredFormat, color, arg)
	}
	return fmt.Sprintf(coloredFormat, arg)
}

// kv is an ansi instance type for Key-Value logging.
type kvLogger struct {
	KeyPainter   painter
	ValuePainter painter
}

// Println print the key with predefined KeyColor and KeyMaxWidth
// and the value with the predefined ValueColor.
func (kv *kvLogger) Println(key interface{}, value interface{}) {
	k, v := kv.ansify(key, value)
	fmt.Printf("%v%v\n", k, v)
}

func (kv *kvLogger) ansify(key interface{}, value interface{}) (string, string) {
	var k, v string

	k = fmt.Sprintf("%-20v", key)

	if kv.KeyPainter == nil {
		kv.KeyPainter = def
	}

	k = kv.KeyPainter(k)

	if kv.ValuePainter != nil {
		v = kv.ValuePainter(value)
	} else {
		v = fmt.Sprint(value)
	}

	return k, v
}

func printLoadHeader() {
	fmt.Printf("%-19v | %-28v | Status\n", "Field name", "Type")
	fmt.Println("--------------------------------------------------------------------------")

}

func printLoadResult(f *reflect.StructField, t reflect.Type, err error) {
	objName := t.Name()
	if f != nil {
		objName = f.Name
	}
	objName = fmt.Sprintf("%-19v", objName)
	objType := fmt.Sprintf("%-28v", t.String())
	if err != nil {
		if err == errOmit {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, err.Error())
		} else if err == errNoConfigurable {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, yellow("-> "+err.Error()))
		} else {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, red("-> "+err.Error()))
		}
	} else {
		fmt.Printf("%s | %s | %s\n", blue(objName), objType, green("<- config loaded"))
	}
}
