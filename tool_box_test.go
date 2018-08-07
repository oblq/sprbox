package sprbox

import (
	"errors"
	"testing"
)

var defaultBoxConfig = Tool{"default config path"}

// Tool is a struct implementing 'configurable' interface.
type Tool struct {
	ConfigPath string
}

// SpareConfig is the 'configurable' interface implementation.
func (c *Tool) SpareConfig(config []string) error {
	LoadConfig(c, config...)
	return nil
}

func (c *Tool) SpareConfigBytes(config []byte) error {
	Unmarshal(config, c)
	return nil
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
	LoadConfig(c, config...)
	return nil
}

func (c *ConfigurableSlicePtr) SpareConfig(config []string) error {
	LoadConfig(c, config...)
	return nil
}

type ConfigurableMap map[string]Tool
type ConfigurableMapPtr map[string]*Tool

func (c *ConfigurableMap) SpareConfig(config []string) error {
	LoadConfig(c, config...)
	return nil
}

func (c *ConfigurableMapPtr) SpareConfig(config []string) error {
	LoadConfig(c, config...)
	return nil
}

type Box struct {
	Tool                  Tool
	PTRTool               *Tool
	ToolNoConfigurable    ToolNoConfigurable
	PTRToolNoConfigurable *ToolNoConfigurable

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

	createJSON(defaultBoxConfig, "Tool.json", t)
	createTOML(defaultBoxConfig, "PTRTool.toml", t)

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
	if test.Tool.ConfigPath != defaultBoxConfig.ConfigPath {
		t.Error("test.Tool.ConfigPath is empty")
	}
	if len(test.PTRTool.ConfigPath) == 0 {
		t.Error("test.PTRTool.ConfigPath is empty")
	}
	if len(test.ToolNoConfigurable.ConfigPath) > 0 {
		t.Error("test.ToolNoConfigurable.ConfigPath:", test.ToolNoConfigurable.ConfigPath)
	}
	if len(test.ToolNoConfigurable.ConfigPath) > 0 {
		t.Error("test.PTRToolNoConfigurable.ConfigPath:", test.PTRToolNoConfigurable.ConfigPath)
	}

	if len(test.ToolSlice[0].ConfigPath) == 0 {
		t.Error("test.ToolSlice.ConfigPath is empty")
	}
	if len(test.ToolSlicePTR[0].ConfigPath) == 0 {
		t.Error("test.ToolSlicePTR.ConfigPath is empty")
	}
	if len((*test.PTRToolSlice)[0].ConfigPath) == 0 {
		t.Error("test.PTRToolSlice.ConfigPath is empty")
	}
	if len(test.ToolMap["test1"].ConfigPath) == 0 {
		t.Error("test.ToolMap.ConfigPath is empty")
	}
	if len(test.ToolMapPTR["test1"].ConfigPath) == 0 {
		t.Error("test.ToolMapPTR.ConfigPath is empty")
	}
	if len((*test.PTRToolMap)["test1"].ConfigPath) == 0 {
		t.Error("test.PTRToolMap.ConfigPath is empty")
	}
	SetDebug(false)
}

type BoxError struct {
	ToolError ToolError
}

func TestBoxError(t *testing.T) {
	createYAML(defaultBoxConfig, "ToolError.yaml", t)
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
	createYAML(defaultBoxConfig, "PTRToolError.yml", t)
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

	createJSON(defaultBoxConfig, "Tool1.json", t)
	createTOML(defaultBoxConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test1 BoxNil
	if err := LoadToolBox(&test1, configPath); err != nil {
		t.Error(err)
	}
	if len(test1.Tool1.ConfigPath) == 0 {
		t.Error("test1.Tool1.ConfigPath:", test1.Tool1.ConfigPath)
	}
	if len(test1.Tool2.ConfigPath) == 0 {
		t.Error("test1.Tool2.ConfigPath:", test1.Tool2.ConfigPath)
	}

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
	if len(test3.Tool1.ConfigPath) == 0 {
		t.Error("test3.Tool1.ConfigPath:", test3.Tool1.ConfigPath)
	}
	if len(test3.Tool2.ConfigPath) == 0 {
		t.Error("test3.Tool2.ConfigPath:", test3.Tool2.ConfigPath)
	}

	SetColoredLogs(true)
}

type BoxConfigFiles struct {
	Tool1 Tool
	Tool2 Tool
	Tool3 *Tool
	Tool4 Tool
}

func TestConfigFiles(t *testing.T) {
	createYAML(defaultBoxConfig, "Tool1.yml", t)
	createJSON(defaultBoxConfig, "Tool3.json", t)
	createTOML(defaultBoxConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test BoxConfigFiles
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}
	if len(test.Tool1.ConfigPath) == 0 {
		t.Error("test.Tool1.ConfigPath:", test.Tool1.ConfigPath)
	}
	if len(test.Tool2.ConfigPath) == 0 {
		t.Error("test.Tool2.ConfigPath:", test.Tool2.ConfigPath)
	}
	if len(test.Tool3.ConfigPath) == 0 {
		t.Error("test.Tool3.ConfigPath:", test.Tool3.ConfigPath)
	}
	if len(test.Tool4.ConfigPath) > 0 {
		t.Error("test.Tool4.ConfigPath:", test.Tool4.ConfigPath)
	}
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

	devConfig := defaultBoxConfig
	devpath := "dev"
	devConfig.ConfigPath = devpath

	createYAML(devConfig, "Tool7.development.yml", t)
	createYAML(defaultBoxConfig, "Tool1.yml", t)
	createYAML(defaultBoxConfig, "test.yml", t)
	createJSON(devConfig, "tool8.development.json", t)
	createTOML(defaultBoxConfig, "Tool2.toml", t)
	defer removeConfigFiles(t)

	var test BoxTags
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}
	if test.Tool1.ConfigPath != defaultBoxConfig.ConfigPath {
		t.Error("test.Tool1.ConfigPath:", test.Tool1.ConfigPath)
	}
	if test.Tool2.ConfigPath == defaultBoxConfig.ConfigPath {
		t.Error("test.Tool2.ConfigPath:", test.Tool2.ConfigPath)
	}
	if test.Tool3.ConfigPath != defaultBoxConfig.ConfigPath {
		t.Error("test.Tool3.ConfigPath:", test.Tool3.ConfigPath)
	}
	if len(test.Tool5.ConfigPath) > 0 {
		t.Error("test.Tool5.ConfigPath:", test.Tool5.ConfigPath)
	}
	if test.Tool6.ConfigPath == defaultBoxConfig.ConfigPath {
		t.Error("test.Tool6 not nil", test.Tool6)
	}
	if test.Tool7.ConfigPath != devpath {
		t.Error("test.Tool7.ConfigPath:", test.Tool7.ConfigPath)
	}
	if test.Tool8.ConfigPath != devpath {
		t.Error("test.Tool8.ConfigPath:", test.Tool8.ConfigPath)
	}
}

type BoxAfterConfig struct {
	Tool1 Tool
	Tool2 Tool `sprbox:"omit"`
}

func TestBoxAfterConfig(t *testing.T) {
	createYAML(defaultBoxConfig, "Tool1.yml", t)
	defer removeConfigFiles(t)

	tString := "must remain the same"
	test := BoxAfterConfig{}
	test.Tool2 = Tool{ConfigPath: tString}
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}
	if len(test.Tool1.ConfigPath) == 0 {
		t.Error("test1.ConfigPath:", test.Tool1.ConfigPath)
	}
	if test.Tool2.ConfigPath != tString {
		t.Error("test2.ConfigPath:", test.Tool2.ConfigPath)
	}
}
