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
	*workerful.Workerful
}

func (wp *Workerful) SpareConfig(configFiles []string) (err error) {
	var cfg workerful.Config
	err = sprbox.LoadConfig(&cfg, configFiles...)
	wp.Workerful = workerful.New("", &cfg)
	return
}

func (wp *Workerful) SpareConfigBytes(configBytes []byte) (err error) {
	var cfg workerful.Config
	err = sprbox.Unmarshal(configBytes, &cfg)
	wp.Workerful = workerful.New("", &cfg)
	return
}
