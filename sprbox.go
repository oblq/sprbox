// Package sprbox is an agnostic config parser
// (supporting YAML, TOML, JSON and Environment vars)
// and a toolbox factory with automatic configuration
// based on your build environment.
package sprbox

import (
	"fmt"
	"os"
	"path/filepath"

	"encoding/json"
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
	// struct field tag key
	sftKey = "sprbox"
)

var (
	// Debug will print some useful info for debug.
	debug = false

	// ColoredLog enable or disable colors in console.
	coloredLogs = true

	// FileSearchCaseSensitive determine config files search mode.
	FileSearchCaseSensitive = true
)

//var testRegexp = regexp.MustCompile(`_test|(\.test$)`)

func init() {
	// automatic debug during tests
	//if testRegexp.MatchString(os.Args[0]) {
	//	Debug()
	//}
}

// ColoredLogs turn on/off colors in console.
func ColoredLogs(colored bool) {
	coloredLogs = colored
}

// SetDebug will print detailed logs in console.
func SetDebug(enabled bool) {
	debug = enabled
}

// PrintInfo print some useful info about
// the environment and git.
func PrintInfo() {
	version := ""
	sprboxRepo := NewRepository(filepath.Join(os.Getenv("GOPATH"), "/src/github.com/oblq/sprbox"))
	if sprboxRepo.Error == nil {
		version = "v" + sprboxRepo.Tag + "(" + sprboxRepo.Build + ")"
	}
	fmt.Printf(darkGrey(banner), version)

	Env().PrintInfo()
	VCS.PrintInfo()
}

func debugPrintf(format string, args ...interface{}) {
	if debug {
		fmt.Printf(format, args...)
	}
}

func dump(v interface{}) string {
	//d, _ := yaml.Marshal(v)
	//return string(d)
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
