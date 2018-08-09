package app

import (
	"github.com/oblq/sprbox"
	"github.com/oblq/sprbox/common/services"
)

// ToolBox is the struct to initialize with sprbox.
// It contains pluggable libraries, implementing the
// 'configurable' interface: `func SpareConfig([]byte) error`.
type ToolBox struct {
	// By default sprbox will look for a file named like the
	// struct field name (Services.*, case sensitive).
	Services services.ServicesMap

	// MediaProcessing does not implement the 'configurable' interface
	// so it will be traversed recursively.
	// Recursion only stop when no more embedded elements are found
	// or when a 'configurable' element is found instead.
	// 'configurable' elements will not be traversed.
	MediaProcessing struct {
		// Optionally pass one or more config file name in the tag,
		// file extension can be omitted.
		Pictures services.Service `sprbox:"MPPict1|MPPict2"`
		Videos   services.Service // sprbox will look for ./config/Videos.* here.
	}

	WP Workerful
	// Workerful implement the 'configurableInCollections' interface,
	// so it can be loaded also directly inside slices or maps using a single config file.
	WPS []Workerful

	// Optionally add the 'omit' value so sprbox will skip that field.
	OmittedTool Tool `sprbox:"omit"`
}

// Shared is the app toolbox, `app.Shared`.
var Shared ToolBox

func init() {
	// set environment manually and debug mode: ------------------------------------------------------------------------

	// Set `testing` build environment manually.
	sprbox.BUILDENV = "staging"

	// optionally turn off colors in logs
	//sprbox.ColoredLogs(false)

	sprbox.PrintInfo()

	// set debug mode
	//sprbox.SetDebug(true)

	// load toolbox: ---------------------------------------------------------------------------------------------------

	sprbox.LoadToolBox(&Shared, "./config")
}
