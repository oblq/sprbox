package sprbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EnvVarKey is the environment variable that
// determine the build environment in sprbox.
const EnvVarKey = "BUILD_ENV"

var (
	// BUILDENV define the current environment.
	// Can be defined by code or, since it is an exported string,
	// can be interpolated with -ldflags at build/run time:
	// 	go build -ldflags "-X github.com/oblq/goms/env.TAG=develop" -v -o ./api_bin ./api
	//
	// If TAG is empty then the environment variable 'BUILD_ENV' will be used.
	//
	// If also the 'BUILD_ENV' environment variable is empty,
	// if you have setup the VCS var, then the git.BranchName will be used.
	// Git-Flow automatic environment selection based on branch name is also supported.
	// Here the default environment RegEx, you can customize them as you want:
	//  - Production 	exps: Exps{"production", "master"}
	//	- Staging 	exps: Exps{"staging", "release/*", "hotfix/*"}
	//	- Testing 	exps: Exps{"testing", "test", "feature/*"}
	//	- Development 	exps: Exps{"development", "develop", "dev"}
	//	- Local 	exps: Exps{"local"}
	BUILDENV = ""

	// VCS is the project version control system.
	// By default it uses the working directory.
	VCS = NewRepository("./")

	privateTAG = ""
)

// Default environment's configuration
var (
	Production  = &Environment{id: "production", exps: []string{"production", "master"}, RunCompiled: true}
	Staging     = &Environment{id: "staging", exps: []string{"staging", "release/*", "hotfix/*"}, RunCompiled: true}
	Testing     = &Environment{id: "testing", exps: []string{"testing", "test", "feature/*"}, RunCompiled: false}
	Development = &Environment{id: "development", exps: []string{"development", "develop", "dev"}, RunCompiled: false}
	Local       = &Environment{id: "local", exps: []string{"local"}, RunCompiled: false}
)

func init() {
	Production.compileExps()
	Staging.compileExps()
	Testing.compileExps()
	Development.compileExps()
	Local.compileExps()
}

var testingRegexp = regexp.MustCompile(`_test|(\.test$)|_Test`)
var inferredBy string

func loadTag() {
	if len(BUILDENV) > 0 {
		privateTAG = BUILDENV
		inferredBy = fmt.Sprintf("'%s', inferred from 'BUILDENV' var, set manually.", privateTAG)
		return
	} else if privateTAG = os.Getenv(EnvVarKey); len(privateTAG) > 0 {
		inferredBy = fmt.Sprintf("'%s', inferred from '%s' environment variable.", privateTAG, EnvVarKey)
		return
	} else if VCS != nil {
		if VCS.Error == nil {
			privateTAG = VCS.BranchName
			inferredBy = fmt.Sprintf("<empty>, inferred from git.BranchName (%s).", VCS.BranchName)
			return
		}
	} else if testingRegexp.MatchString(os.Args[0]) {
		privateTAG = Testing.ID()
		inferredBy = fmt.Sprintf("'%s', inferred from the running file name (%s).", privateTAG, os.Args[0])
		return
	}

	inferredBy = "<empty>, default environment is 'local'."
}

// Env returns the current selected environment by
// matching the privateTAG variable against the environments RegEx.
func Env() *Environment {
	loadTag()
	switch {
	case Production.MatchTag(privateTAG):
		return Production
	case Staging.MatchTag(privateTAG):
		return Staging
	case Testing.MatchTag(privateTAG):
		return Testing
	case Development.MatchTag(privateTAG):
		return Development
	case Local.MatchTag(privateTAG):
		return Local
	default:
		return Local
	}
}

// EnvSubDir returns <path>/<environment>
func EnvSubDir(path string) string {
	return filepath.Join(path, Env().ID())
}

// CompiledPath() returns the path base if RunCompiled == true
// for the environment in use so that static files can
// stay side by side with the executable
// while it is possible to have a different location when the
// program is launched with `go run`.
// This allow to manage multiple packages in one project during development,
// for instance using a config path in the parent dir, side by side with
// the packages, while having the same config folder side by side with
// the executable where needed.
//
// Can be used in:
//  sprbox.LoadToolBox(&myToolBox, sprbox.CompiledPath("../config"))
//
// Example:
//  sprbox.Development.RunCompiled = false
//  sprbox.BUILDENV = sprbox.Development.ID()
//  sprbox.CompiledPath("../static_files/config") // -> "../static_files/config"
//
//  sprbox.Development.RunCompiled = true
//  sprbox.BUILDENV = sprbox.Development.ID()
//  sprbox.CompiledPath("../static_files/config") // -> "config"
//
// By default only Production and Staging environments have RunCompiled = true.
func CompiledPath(path string) string {
	if Env().RunCompiled {
		return filepath.Base(path)
	}
	return path
}

// Environment struct.
type Environment struct {
	id     string
	exps   []string
	regexp *regexp.Regexp

	// RunCompiled true means that the program run from
	// a precompiled binary for that environment.
	// CompiledPath() returns the path base if RunCompiled == true
	// so that static files can stay side by side with the executable
	// while it is possible to have a different location when the
	// program is launched with `go run`.
	//
	// By default only Production and Staging environments have RunCompiled = true.
	RunCompiled bool
}

// MatchTag check if the passed tag match that environment,
// a tag may be the branch name or the machine hostname or whatever you want.
func (e *Environment) MatchTag(tag string) bool {
	return e.regexp.MatchString(tag)
}

// AppendExp add a regular expression to match that environment.
func (e *Environment) AppendExp(exp string) {
	e.exps = append(e.exps, exp)
	e.compileExps()
}

// SetExps set regular expressions to match that environment.
func (e *Environment) SetExps(exps []string) {
	e.exps = exps
	e.compileExps()
}

func (e *Environment) compileExps() {
	regex := "(" + strings.Join(e.exps, ")|(") + ")"
	e.regexp = regexp.MustCompile(regex)
}

// ID returns the environment id,
// which are also a valid tag for the current environment.
func (e *Environment) ID() string {
	return e.id
}

// Info return some environment info.
func (e *Environment) Info() string {
	return fmt.Sprintf("%s - tag: %s\n", strings.ToUpper(e.ID()), inferredBy)
}

// PrintInfo print some environment info in console.
func (e *Environment) PrintInfo() {
	info := fmt.Sprintf("%s - tag: %s\n", green(strings.ToUpper(e.ID())), inferredBy)
	envLog := kvLogger{}
	envLog.Println("Environment:", info)
	//envLog.Println("Config path:", ansi.Green(ConfigPathByEnv(configPath))+"\n")
}
