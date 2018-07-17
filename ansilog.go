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
	//defaultFG color = "39"

	//blackCol color = "97" // inverted with white
	redCol     color = "31"
	greenCol   color = "32"
	yellowCol  color = "33"
	blueCol    color = "34"
	magentaCol color = "35"
	//cyanCol         color = "36"
	//lightGreyCol    color = "37"
	darkGreyCol color = "90"
	//lightRedCol     color = "91"
	//lightGreenCol   color = "92"
	//lightYellowCol  color = "93"
	//lightBlueCol    color = "94"
	//lightMagentaCol color = "95"
	//lightCyanCol    color = "96"
	//whiteCol        color = "30" // inverted with black

	esc   = "\033["
	clear = "\033[0m"
)

type painter func(interface{}) string

// Black return the argument as a color escaped string
//func black(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), blackCol)
//}

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

// Cyan return the argument as a color escaped string
//func cyan(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), cyanCol)
//}

// LightGrey return the argument as a color escaped string
//func lightGrey(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightGreyCol)
//}

// DarkGrey return the argument as a color escaped string
func darkGrey(arg interface{}) string {
	return colorize(fmt.Sprint(arg), darkGreyCol)
}

// LightRed return the argument as a color escaped string
//func lightRed(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightRedCol)
//}

// LightGreen return the argument as a color escaped string
//func lightGreen(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightGreenCol)
//}

// LightYellow return the argument as a color escaped string
//func lightYellow(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightYellowCol)
//}

// LightBlue return the argument as a color escaped string
//func lightBlue(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightBlueCol)
//}

// LightMagenta return the argument as a color escaped string
//func lightMagenta(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightMagentaCol)
//}

// LightCyan return the argument as a color escaped string
//func lightCyan(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), lightCyanCol)
//}

// White return the argument as a color escaped string
//func white(arg interface{}) string {
//	return colorize(fmt.Sprint(arg), whiteCol)
//}

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

// Print print the key with predefined KeyColor and width
// and the value with the predefined ValueColor.
func (kv *kvLogger) Print(key interface{}, value interface{}) {
	k, v := kv.ansify(key, value)
	fmt.Printf("%v%v", k, v)
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

	if kv.KeyPainter != nil {
		k = kv.KeyPainter(k)
	}

	if kv.ValuePainter != nil {
		v = kv.ValuePainter(value)
	} else {
		v = fmt.Sprint(value)
	}

	return k, v
}

func printLoadHeader() {
	fmt.Printf("%-16v | %-28v | Result\n", "Field name", "Type")
	fmt.Println("----------------------------------------------------------------")

}

func printLoadResult(f *reflect.StructField, t reflect.Type, err error) {
	objName := t.Name()
	if f != nil {
		objName = f.Name
	}
	objName = fmt.Sprintf("%-16v", objName)
	objType := fmt.Sprintf("%-28v", t.String())
	if err != nil {
		if err == errOmit {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, err.Error())
		} else if err == errNoBoxable {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, yellow("-> "+err.Error()))
		} else {
			fmt.Printf("%s | %s | %s\n", blue(objName), objType, red("-> "+err.Error()))
		}
	} else {
		fmt.Printf("%s | %s | %s\n", blue(objName), objType, green("<- loaded"))
	}
}
