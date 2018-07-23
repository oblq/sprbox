package toolbox

import (
	"github.com/oblq/sprbox"
	"github.com/oblq/workerful"
)

// ToolBox is the struct to initialize with sprbox.
// It contains pluggable libraries, implementing the
// 'configurable' interface: func SBConfig(configPath string) error.
type ToolBox struct {
	WPool workerful.Workerful `config:"workerpool.yml"`

	// By default sprbox will look for a file named like the
	// struct field name (ATool.yml, case sensitive).
	ATool    Tool
	AToolPtr *Tool `config:"ATool.yml"`

	// Optionally pass a config file name in the tag.
	ATool2 Tool

	NotConfigurable struct{ Text string }

	// Optionally add the 'omit' value so sprbox will skip that field.
	AnOmittedTool Tool `omit:"true"`
}

// App is the app toolbox
var App ToolBox

func init() {
	// Set `local` build environment manually.
	sprbox.BUILDENV = "local"

	// Print some useful info.
	sprbox.PrintInfo(false)

	err := sprbox.Load(&App, "example/config")
	if err != nil {
		panic(err)
	}

	App.NotConfigurable.Text = "some text..."

	// From here on you can grab your libs, fully initialized and configured.
}
