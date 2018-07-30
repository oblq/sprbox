package app

import (
	"github.com/oblq/sprbox"
	"github.com/oblq/workerful"
)

// Example:
// How to embed a third-party package in sprbox.
//
// workerful.Workerful already implement `func SpareConfig([]byte) error`
// `configurable` interface, but that's only for demonstration.
//
// Workerful embed workerful.Workerful adding
// the `configurable` interface to it.

type Workerful struct {
	workerful.Workerful
}

func (wp *Workerful) SpareConfig(configData []byte) (err error) {
	// since the config file format is known here
	// you can use yaml, toml or json unmarshaler directly.
	// sprbox.Unmarshal() will recognize any of those formats
	// and will process sprbox flags eventually.
	var cfg workerful.Config
	err = sprbox.Unmarshal(configData, &cfg)
	wp.Workerful = *workerful.New("", &cfg)
	return
}
