package main

import (
	"fmt"
	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/oblq/sprbox"
	"github.com/oblq/workerful"
	"gopkg.in/yaml.v2"
)

// Tool is a struct example that .
type Tool struct {
	Text string `yaml:"text"`
}

// SBConfig is the https://github.com/oblq/sprbox 'configurable' interface implementation.
func (t *Tool) SBConfig(configPath string) error {
	if compsConfigFile, err := ioutil.ReadFile(configPath); err != nil {
		return fmt.Errorf("wrong config path: %s", err.Error())
	} else if err = yaml.Unmarshal(compsConfigFile, t); err != nil {
		return fmt.Errorf("can't unmarshal config file: %s", err.Error())
	}
	return nil
}

func (t *Tool) getText() string {
	return t.Text
}

// AppToolBox is the struct to initialize with sprbox.
// It contains pluggable libraries, implementing the
// 'configurable' interface: func SBConfig(configPath string) error.
type AppToolBox struct {
	WPool workerful.Workerful `sprbox:"workerpool.yml"`

	// By default sprbox will look for a file named like the
	// struct field name (ATool.yml, case insensitive).
	ATool    Tool
	AToolPtr *Tool `sprbox:"atool.yml"`

	// Optionally pass a config file name in the tag.
	ATool2 Tool

	NotConfigurable struct{ Text string }

	// Optionally add the 'omit' value so sprbox will skip that field.
	AnOmittedTool Tool `sprbox:"omit"`
}

// App is the app toolbox
var App AppToolBox

func init() {
	// The project must contain a config folder for any
	// build environment you need to use.
	// Sprbox support the 5 standard environment,
	// the right one is chosen by matching a RegEx and
	// the tag to match will be taken from three parameters,
	// in that precise order:
	// 1. the BUILDENV var in sprbox package (sprbox.BUILDENV).
	// 2. the environment variable `BUILD_ENV` (os.GetEnv("BUILD_ENV")).
	// 3. the branch name. Git-Flow supported.
	//    (you must pass the git repo path: sprbox.VCS = sprbox.NewRepository("path/to/repo")).
	//
	// Set `local` build environment manually.
	// This is only a tag to match against the environments RegEx.
	// You can define the RegEx patterns per environment,
	// the default patterns are:
	//  - Production 	exps: Exps{"production", "master"}
	//	- Staging 		exps: Exps{"staging", "release/*", "hotfix/*"}
	//	- Testing 		exps: Exps{"testing", "test", "feature/*"}
	//	- Development 	exps: Exps{"development", "develop", "dev"}
	//	- Local 		exps: Exps{"local", ""}
	//sprbox.BUILDENV = "local"

	// Optionally set the repository path.
	// In that case we set the environment
	// manually, which take precedence on Git Branch name,
	// Anyway having the repo set will print also git info in console.
	//sprbox.VCS = sprbox.NewRepository("./")

	// Print some useful info.
	sprbox.PrintInfo(false)

	err := sprbox.Load(&App, "example/config")
	if err != nil {
		panic(err)
	}

	App.NotConfigurable.Text = "some text..."

	// From here on you can grab your libs, fully initialized and configured.
}

// CEC is a custom echo context
type CEC struct {
	echo.Context
	App *AppToolBox
}

// EchoSprBox provides the AppBox (inherited from echo.Context) to echo.
// This middleware should be registered before any other.
func EchoSprBox(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// pass the pointer to 'app'
		embeddedBox := &CEC{Context: c, App: &App}
		return h(embeddedBox)
	}
}
