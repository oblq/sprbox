package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Tool is a struct example that .
type Tool struct {
	Text string `yaml:"text"`
}

// Go2Box is the https://github.com/oblq/sprbox 'boxable' interface implementation.
func (t *Tool) Go2Box(configPath string) error {
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
