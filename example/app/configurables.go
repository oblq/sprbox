package app

import (
	"github.com/oblq/sprbox"
)

// Example:
// How to use your package in sprbox.
//
// Tool is a struct example that implement
// the `configurable` interface natively.

type Tool struct {
	Text string `yaml:"text"`
}

// SpareConfig is the 'configurable' interface implementation.
func (t *Tool) SpareConfig(configData []byte) (err error) {
	// since the config file format is known here
	// you can use yaml, toml or json unmarshaler directly.
	// sprbox.Unmarshal() will recognize any of those formats
	// and will process sprbox flags.
	err = sprbox.Unmarshal(configData, t)
	return
}

// GetText returns the text stored in Tool
func (t *Tool) GetText() string {
	return t.Text
}
