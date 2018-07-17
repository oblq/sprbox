package sprbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EnvVarKey define the name of the environment variable to look for.
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
	// if you have called SetUpGit(path), then the git.BranchName will be used.
	// Git-Flow automatic environment selection based on branch name is also supported.
	// Here the default environment RegEx, you can customize them as you want:
	//  - Production 	exps: Exps{"production", "master"}
	//	- Staging 		exps: Exps{"staging", "release/*", "hotfix/*"}
	//	- Testing 		exps: Exps{"testing", "test", "feature/*"}
	//	- Development 	exps: Exps{"development", "develop", "dev"}
	//	- Local 		exps: Exps{"local"}
	BUILDENV = ""

	// VCS is the project version control system, by default is nil.
	VCS = NewRepository("./")

	privateTAG = ""

	configPath = ""

	envLog = kvLogger{}
)

// Default environment's configuration
var (
	Production  = &Environment{id: "production", exps: []string{"production", "master"}}
	Staging     = &Environment{id: "staging", exps: []string{"staging", "release/*", "hotfix/*"}}
	Testing     = &Environment{id: "testing", exps: []string{"testing", "test", "feature/*"}}
	Development = &Environment{id: "development", exps: []string{"development", "develop", "dev"}}
	Local       = &Environment{id: "local", exps: []string{"local"}}
)

func init() {
	Production.compileExps()
	Staging.compileExps()
	Testing.compileExps()
	Development.compileExps()
	Local.compileExps()

	loadTag()
}

var inferredBy string

func loadTag() {
	if len(BUILDENV) > 0 {
		privateTAG = BUILDENV
		inferredBy = ", inferred from 'BUILDENV' var, set manually."
		return
	} else if privateTAG = os.Getenv(EnvVarKey); len(privateTAG) > 0 {
		inferredBy = fmt.Sprintf(", inferred from '%s' environment variable.", EnvVarKey)
		return
	} else if VCS != nil {
		if VCS.Error == nil {
			privateTAG = VCS.BranchName
			inferredBy = "'', inferred from git.BranchName."
			return
		}
	}

	inferredBy = "'', default environment is 'local'."
}

// Env returns the current selected environment by
// matching the ENVTag variable against the environments RegEx (Exps).
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

// ConfigPathByEnv returns <path>/<environment>
func ConfigPathByEnv(path string) string {
	return filepath.Join(path, Env().String())
}

// Environment struct.
type Environment struct {
	id     string
	exps   []string
	regexp *regexp.Regexp
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

// String returns the environment name,
// which are also a valid tag for the current environment.
func (e *Environment) String() string {
	return e.id
}

// Info return some environment info.
func (e *Environment) Info() string {
	return fmt.Sprintf("%s - tag: %s%s\n", strings.ToUpper(e.String()), BUILDENV, inferredBy)
}

// PrintInfo print some environment info in console.
func (e *Environment) PrintInfo() {
	info := fmt.Sprintf("%s - tag: %s%s\n", green(strings.ToUpper(e.String())), BUILDENV, inferredBy)
	envLog.Println("Environment:", info)
	//envLog.Println("Config path:", ansi.Green(ConfigPathByEnv(configPath))+"\n")
}
