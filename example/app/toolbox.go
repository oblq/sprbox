package app

import (
	"github.com/oblq/sprbox"
)

// ToolBox is the struct to initialize with sprbox.
// It contains pluggable libraries, implementing the
// 'configurable' interface: `func SBConfig([]byte) error`.
type ToolBox struct {
	// By default sprbox will look for a file named like the
	// struct field name (WPool.*, case sensitive).
	WPool        Workerful
	WPoolOmitted Workerful `sprbox:"omit"`

	Tool1    Tool
	Tool1Ptr *Tool `sprbox:"Tool1.yml"`

	// Optionally pass a config file name in the tag.
	Tool2 Tool

	NotConfigurable *struct {
		Text string
	}

	// Optionally add the 'omit' value so sprbox will skip that field.
	OmittedTool Tool `sprbox:"omit"`
}

// App is the app toolbox
var Shared ToolBox

func init() {
	// set environment manually and debug mode: ------------------------------------------------------------------------

	// Set `testing` build environment manually.
	sprbox.BUILDENV = "testing"

	// optionally turn off colors in logs
	//sprbox.ColoredLogs(false)

	// set debug mode
	sprbox.SetDebug(true)

	// load toolbox: ---------------------------------------------------------------------------------------------------

	err := sprbox.LoadToolBox(&Shared, "./config")
	if err != nil {
		panic(err)
	}

	// initialized but not configured...
	Shared.NotConfigurable.Text = "some text..."

	// From here on you can grab your libs, fully initialized and configured.
}
