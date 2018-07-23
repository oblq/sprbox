package toolbox

import (
	"fmt"
	"io/ioutil"

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

// GetText returns the text stored in Tool
func (t *Tool) GetText() string {
	return t.Text
}
