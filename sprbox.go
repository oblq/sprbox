// Package sprbox is an agnostic config parser
// (supporting YAML, TOML, JSON and Environment vars)
// and a toolbox factory with automatic configuration
// based on your build environment.
package sprbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

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

var testRegexp = regexp.MustCompile(`_test|(\.test$)`)

func init() {
	// automatic debug during tests
	if testRegexp.MatchString(os.Args[0]) {
		Debug()
	}
}

// ColoredLogs turn on/off colors in console.
func ColoredLogs(colored bool) {
	coloredLogs = colored
}

// Debug will print detailed logs in console.
func Debug() {
	debug = true

	version := ""
	sprboxRepo := NewRepository(filepath.Join(os.Getenv("GOPATH"), "/src/github.com/oblq/sprbox"))
	if sprboxRepo.Error == nil {
		version = "v" + sprboxRepo.Tag + "(" + sprboxRepo.Build + ")"
	} else {
		println(sprboxRepo.Error.Error())
	}
	fmt.Printf(darkGrey(banner), version)

	PrintInfo()
}

// PrintInfo print some useful info about
// the environment and git.
func PrintInfo() {
	fmt.Printf("\n")
	Env().PrintInfo()
	VCS.PrintInfo()
}

func debugPrintf(format string, args ...interface{}) {
	if debug {
		fmt.Printf(format, args...)
	}
}

func prettyPrinted(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
