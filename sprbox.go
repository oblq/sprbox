// Package sprbox is an agnostic config parser
// (supporting YAML, TOML, JSON and Environment vars)
// and a toolbox factory with automatic configuration
// based on your build environment.
package sprbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

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

func init() {
	//coloredLogs = isTerminal(os.Stdout.Fd())
}

// isTerminal return true if the file descriptor is terminal.
func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

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
	// To marshal directly with yaml produce a panic with unexported fields

	jd, err := json.Marshal(v)
	if err != nil {
		//fmt.Printf("dump err on %+v: %v\n", v, err)
		return fmt.Sprintf("%+v", v)
	}

	// Convert the JSON to an object.
	var jsonObj interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshalling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	err = yaml.Unmarshal(jd, &jsonObj)
	if err != nil {
		//fmt.Printf("dump err on %+v: %v\n", v, err)
		return fmt.Sprintf("%+v", v)
	}

	// Marshal this object into YAML.
	yd, err := yaml.Marshal(jsonObj)
	if err != nil {
		//fmt.Printf("dump err on %+v: %v\n", v, err)
		return fmt.Sprintf("%+v", v)
	}
	return string(yd)
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

func GetInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "\n%s%s\n%s\n", banner, Env().Info(), VCS.Info())
}

// SetColoredLogs toggle colors in console.
func SetColoredLogs(colored bool) {
	coloredLogs = colored
}

// SetFileSearchCaseSensitive toggle case sensitive cinfig files search.
func SetFileSearchCaseSensitive(caseSensitive bool) {
	fileSearchCaseSensitive = caseSensitive
}
