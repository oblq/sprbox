// Package sprbox is an agnostic config parser
// (supporting YAML, TOML, JSON and Environment vars)
// and a toolbox factory with automatic configuration
// based on your build environment.
package sprbox

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
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
	debug   = false
	verbose = false

	// coloredLogs enable or disable colors in console.
	coloredLogs = true

	// fileSearchCaseSensitive determine config files search mode.
	fileSearchCaseSensitive = true
)

// SetDebug will print detailed logs in console.
func SetDebug(enabled bool) {
	debug = enabled
}

func debugPrintf(format string, args ...interface{}) {
	if debug {
		fmt.Printf(format, args...)
	}
}

// verbose mode can only be set here
func verbosePrintf(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format, args...)
	}
}

func dump(v interface{}) string {
	d, err := yaml.Marshal(v)
	if err != nil {
		fmt.Println(err)
	}
	return string(d)
	//b, _ := json.MarshalIndent(v, "", "  ")
	//return string(b)+"\n"
}

// PrintInfo print some useful info about the environment and git.
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

// SetColoredLogs toggle colors in console.
func SetColoredLogs(colored bool) {
	coloredLogs = colored
}

// SetFileSearchCaseSensitive toggle case sensitive cinfig files search.
func SetFileSearchCaseSensitive(caseSensitive bool) {
	fileSearchCaseSensitive = caseSensitive
}
