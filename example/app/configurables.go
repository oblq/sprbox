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
func (t *Tool) SpareConfig(configFiles []string) (err error) {
	err = sprbox.LoadConfig(t, configFiles...)
	return
}

// GetText returns the text stored in Tool
func (t *Tool) GetText() string {
	return t.Text
}
