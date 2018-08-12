package sprbox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultToolConfig = Tool{"default config path"}

// Tool is a struct implementing 'configurable' interface.
type Tool struct {
	ConfigPath string
}

// SpareConfig is the 'configurable' interface implementation.
func (c *Tool) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

func (c *Tool) SpareConfigBytes(config []byte) error {
	return Unmarshal(config, c)
}

// ToolError is a struct implementing 'configurable' interface.
type ToolError struct {
	ConfigPath string
}

// SpareConfig is the 'configurable' interface implementation.
func (c *ToolError) SpareConfig(config []string) error {
	return errors.New("fake error for test")
}

// ToolNoConfigurable is a struct implementing 'configurable' interface.
type ToolNoConfigurable struct {
	ConfigPath string
}

type ConfigurableSlice []Tool
type ConfigurableSlicePtr []*Tool

func (c *ConfigurableSlice) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

func (c *ConfigurableSlicePtr) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

type ConfigurableMap map[string]Tool
type ConfigurableMapPtr map[string]*Tool

func (c *ConfigurableMap) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

func (c *ConfigurableMapPtr) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

type Box struct {
	Tool                  Tool
	PTRTool               *Tool
	ToolNoConfigurable    ToolNoConfigurable
	PTRToolNoConfigurable *ToolNoConfigurable

	SubBox struct {
		Tool1 Tool
	}

	ToolSlice    []Tool
	ToolSlicePTR []*Tool
	PTRToolSlice *[]Tool
	ToolMap      map[string]Tool
	ToolMapPTR   map[string]*Tool
	PTRToolMap   *map[string]Tool

	ConSlice    ConfigurableSlice     `sprbox:"ToolSlice.yml"`
	ConSlicePtr *ConfigurableSlicePtr `sprbox:"ToolSlice.yml"`
	ConMap      ConfigurableMap       `sprbox:"ToolMap.yml"`
	ConMapPtr   *ConfigurableMapPtr   `sprbox:"ToolMap.yml"`

	ConSliceOmit ConfigurableSlice `sprbox:"omit"`
	ConMapOmit   ConfigurableMap   `sprbox:"omit"`
}

func TestBox(t *testing.T) {
	SetDebug(true)

	createJSON(defaultToolConfig, "Tool.json", t)
	createTOML(defaultToolConfig, "PTRTool.toml", t)
	createYAML(defaultToolConfig, "SubBox/Tool1.yaml", t)

	ts := []Tool{
		Tool{"test1"},
		Tool{"test2"},
	}
	createYAML(ts, "ToolSlice.yml", t)
	createYAML(ts, "PTRToolSlice.yml", t)

	tsptr := []*Tool{
		&Tool{"test1"},
		&Tool{"test2"},
	}
	createJSON(tsptr, "ToolSlicePTR.json", t)

	tm := map[string]Tool{
		"test1": Tool{"test1"},
		"test2": Tool{"test2"},
	}
	createYAML(tm, "ToolMap.yml", t)
	createTOML(tm, "PTRToolMap.toml", t)

	tmptr := map[string]*Tool{
		"test1": &Tool{"test1"},
		"test2": &Tool{"test2"},
	}
	createJSON(tmptr, "ToolMapPTR.json", t)

	defer removeConfigFiles(t)

	SetFileSearchCaseSensitive(true)

	PrintInfo()

	var test Box
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}

	assert.Equal(t, defaultToolConfig.ConfigPath, test.SubBox.Tool1.ConfigPath, "subBox not correctly loaded")
	assert.Equal(t, defaultToolConfig.ConfigPath, test.Tool.ConfigPath, "test.Tool.ConfigPath is empty")
	assert.NotEqual(t, 0, len(test.PTRTool.ConfigPath), "test.PTRTool.ConfigPath is empty")
	assert.Equal(t, 0, len(test.ToolNoConfigurable.ConfigPath), "test.ToolNoConfigurable.ConfigPath:", test.ToolNoConfigurable.ConfigPath)
	assert.Equal(t, 0, len(test.PTRToolNoConfigurable.ConfigPath), "test.PTRToolNoConfigurable.ConfigPath:", test.PTRToolNoConfigurable.ConfigPath)

	assert.NotEqual(t, 0, len(test.ToolSlice[0].ConfigPath) == 0, "test.ToolSlice.ConfigPath is empty")
	assert.NotEqual(t, 0, len(test.ToolSlicePTR[0].ConfigPath), "test.ToolSlicePTR.ConfigPath is empty")
	assert.NotEqual(t, 0, len((*test.PTRToolSlice)[0].ConfigPath), "test.PTRToolSlice.ConfigPath is empty")
	assert.NotEqual(t, 0, len(test.ToolMap["test1"].ConfigPath), "test.ToolMap.ConfigPath is empty")
	assert.NotEqual(t, 0, len(test.ToolMapPTR["test1"].ConfigPath), "test.ToolMapPTR.ConfigPath is empty")
	assert.NotEqual(t, 0, len((*test.PTRToolMap)["test1"].ConfigPath), "test.PTRToolMap.ConfigPath is empty")

	SetDebug(false)
}

type BoxError struct {
	ToolError ToolError
}

func TestBoxError(t *testing.T) {
	createYAML(defaultToolConfig, "ToolError.yaml", t)
	defer removeConfigFiles(t)

	var test BoxError
	if err := LoadToolBox(&test, configPath); err == nil {
		t.Error(err)
	}
}

type PTRToolError struct {
	PTRToolError *ToolError
}

func TestPTRToolError(t *testing.T) {
	createYAML(defaultToolConfig, "PTRToolError.yml", t)
	defer removeConfigFiles(t)

	var test PTRToolError
	if err := LoadToolBox(&test, configPath); err == nil {
		t.Error(err)
	}
}

func TestInvalidPointer(t *testing.T) {
	var test1 *string
	if err := LoadToolBox(&test1, configPath); err != errInvalidPointer {
		t.Error(err)
	}

	var test2 *Box
	if err := LoadToolBox(test2, configPath); err != errInvalidPointer {
		t.Error(err)
	}
}

type BoxNil struct {
	Tool1 Tool
	Tool2 *Tool
}

func TestNilBox(t *testing.T) {
	SetColoredLogs(false)

	createJSON(defaultToolConfig, "Tool1.json", t)
	createTOML(defaultToolConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test1 BoxNil
	if err := LoadToolBox(&test1, configPath); err != nil {
		t.Error(err)
	}
	assert.NotEqual(t, 0, len(test1.Tool1.ConfigPath), "test1.Tool1.ConfigPath:", test1.Tool1.ConfigPath)
	assert.NotEqual(t, 0, len(test1.Tool2.ConfigPath), "test1.Tool2.ConfigPath:", test1.Tool2.ConfigPath)

	var test2 *BoxNil
	if err := LoadToolBox(test2, configPath); err != nil {
		t.Log(err)
	} else {
		t.Error(err)
	}

	var test3 = &BoxNil{}
	if err := LoadToolBox(test3, configPath); err != nil {
		t.Error(err)
	}
	assert.NotEqual(t, 0, len(test3.Tool1.ConfigPath), "test3.Tool1.ConfigPath:", test3.Tool1.ConfigPath)
	assert.NotEqual(t, 0, len(test3.Tool2.ConfigPath), "test3.Tool2.ConfigPath:", test3.Tool2.ConfigPath)

	SetColoredLogs(true)
}

type BoxConfigFiles struct {
	Tool1 Tool
	Tool2 Tool
	Tool3 *Tool
}

func TestConfigFiles(t *testing.T) {
	createYAML(defaultToolConfig, "Tool1.yml", t)
	createJSON(defaultToolConfig, "Tool3.json", t)
	createTOML(defaultToolConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test BoxConfigFiles
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}
	assert.NotEqual(t, 0, len(test.Tool1.ConfigPath), "test.Tool1.ConfigPath:", test.Tool1.ConfigPath)
	assert.NotEqual(t, 0, len(test.Tool2.ConfigPath), "test.Tool2.ConfigPath:", test.Tool2.ConfigPath)
	assert.NotEqual(t, 0, len(test.Tool3.ConfigPath), "test.Tool3.ConfigPath:", test.Tool3.ConfigPath)
}

func TestConfigFileNotFound(t *testing.T) {
	var test BoxConfigFiles
	assert.Error(t, errConfigFileNotFound, LoadToolBox(&test, configPath), "enexistent config file does not return error")
}

type BoxTags struct {
	Tool1 Tool
	Tool2 Tool  `sprbox:"omit"`
	Tool3 Tool  `sprbox:"test.yml"`
	Tool5 Tool  `sprbox:"test.yml,omit"`
	Tool6 *Tool `sprbox:"omit"`
	Tool7 *Tool
	Tool8 *Tool `sprbox:"tool8"`
}

func TestBoxTags(t *testing.T) {
	BUILDENV = "dev"

	devConfig := defaultToolConfig
	devpath := "dev"
	devConfig.ConfigPath = devpath

	createYAML(devConfig, "Tool7.development.yml", t)
	createYAML(defaultToolConfig, "Tool1.yml", t)
	createYAML(defaultToolConfig, "test.yml", t)
	createJSON(devConfig, "tool8.development.json", t)
	createTOML(defaultToolConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test BoxTags
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}
	assert.Equal(t, defaultToolConfig.ConfigPath, test.Tool1.ConfigPath, "test.Tool1.ConfigPath:", test.Tool1.ConfigPath)
	assert.NotEqual(t, defaultToolConfig.ConfigPath, test.Tool2.ConfigPath, "test.Tool2.ConfigPath:", test.Tool2.ConfigPath)
	assert.Equal(t, defaultToolConfig.ConfigPath, test.Tool3.ConfigPath, "test.Tool3.ConfigPath:", test.Tool3.ConfigPath)
	assert.Equal(t, 0, len(test.Tool5.ConfigPath), "test.Tool5.ConfigPath:", test.Tool5.ConfigPath)
	assert.NotEqual(t, defaultToolConfig.ConfigPath, test.Tool6.ConfigPath, "test.Tool6.ConfigPath:", test.Tool6.ConfigPath)
	assert.Equal(t, devpath, test.Tool7.ConfigPath, "test.Tool7.ConfigPath:", test.Tool7.ConfigPath)
	assert.Equal(t, devpath, test.Tool8.ConfigPath, "test.Tool8.ConfigPath:", test.Tool8.ConfigPath)
}

type BoxAfterConfig struct {
	Tool1 Tool
	Tool2 Tool `sprbox:"omit"`
}

func TestBoxAfterConfig(t *testing.T) {
	createYAML(defaultToolConfig, "Tool1.yml", t)
	defer removeConfigFiles(t)

	tString := "must remain the same"
	test := BoxAfterConfig{}
	test.Tool2 = Tool{ConfigPath: tString}
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}

	assert.NotEqual(t, 0, len(test.Tool1.ConfigPath), "test1.ConfigPath:", test.Tool1.ConfigPath)
	assert.Equal(t, tString, test.Tool2.ConfigPath, "test2.ConfigPath:", test.Tool2.ConfigPath)
}
