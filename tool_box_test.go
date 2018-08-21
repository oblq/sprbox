package sprbox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultToolConfig = ToolConfig{Path: configPath}

type ToolConfig struct {
	Path string
}

// Tool is a struct implementing 'configurable' interface.
type Tool struct {
	Config ToolConfig
}

// SpareConfig is the 'configurable' interface implementation.
func (c *Tool) SpareConfig(config []string) error {
	return LoadConfig(&c.Config, config...)
}

func (c *Tool) SpareConfigBytes(config []byte) error {
	return Unmarshal(config, &c.Config)
}

// ToolError is a struct implementing 'configurable' interface.
type ToolError struct {
	Path string
}

// SpareConfig is the 'configurable' interface implementation.
func (c *ToolError) SpareConfig(config []string) error {
	return errors.New("fake error for test")
}

// ToolNoConfigurable is a struct implementing 'configurable' interface.
type ToolNoConfigurable struct {
	Path string
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

type SubBoxConfigurable struct {
	Path string
	Tool Tool `sprbox:"SubBox/Tool1"`
}

func (c *SubBoxConfigurable) SpareConfig(config []string) error {
	return LoadConfig(c, config...)
}

type Box struct {
	Tool                  Tool
	PTRTool               *Tool
	ToolNoConfigurable    ToolNoConfigurable
	PTRToolNoConfigurable *ToolNoConfigurable

	SubBox struct {
		Tool1 Tool `sprbox:"SubBox/Tool1"`
	}

	SubBoxConfigurable SubBoxConfigurable `sprbox:"Tool"`

	ToolSlice    []Tool
	ToolSlicePTR []*Tool
	PTRToolSlice *[]Tool
	ToolMap      map[string]Tool
	ToolMapPTR   map[string]*Tool
	PTRToolMap   *map[string]Tool

	ConSlice    ConfigurableSlice     `sprbox:"conToolSlice.yml"`
	ConSlicePtr *ConfigurableSlicePtr `sprbox:"conToolSlice.yml"`
	ConMap      ConfigurableMap       `sprbox:"conToolMap.yml"`
	ConMapPtr   *ConfigurableMapPtr   `sprbox:"conToolMap.yml"`

	ConSliceOmit ConfigurableSlice `sprbox:"-"`
	ConMapOmit   ConfigurableMap   `sprbox:"-"`
}

func TestBox(t *testing.T) {
	SetDebug(true)

	createJSON(defaultToolConfig, "Tool.json", t)
	createTOML(defaultToolConfig, "PTRTool.toml", t)
	createYAML(defaultToolConfig, "SubBox/Tool1.yaml", t)

	ts := []ToolConfig{
		ToolConfig{configPath},
		ToolConfig{configPath},
	}
	createYAML(ts, "ToolSlice.yml", t)
	createYAML(ts, "PTRToolSlice.yml", t)

	tsptr := []*ToolConfig{
		&ToolConfig{configPath},
		&ToolConfig{configPath},
	}
	createJSON(tsptr, "ToolSlicePTR.json", t)

	tm := map[string]ToolConfig{
		"test1": ToolConfig{configPath},
		"test2": ToolConfig{configPath},
	}
	createYAML(tm, "ToolMap.yml", t)
	createTOML(tm, "PTRToolMap.toml", t)

	tmptr := map[string]*ToolConfig{
		"test1": &ToolConfig{configPath},
		"test2": &ToolConfig{configPath},
	}
	createJSON(tmptr, "ToolMapPTR.json", t)

	conTS := []Tool{
		Tool{Config: ToolConfig{configPath}},
		Tool{Config: ToolConfig{configPath}},
	}
	createYAML(conTS, "conToolSlice.yml", t)
	createYAML(conTS, "conPTRToolSlice.yml", t)

	conTSptr := []*Tool{
		&Tool{Config: ToolConfig{configPath}},
		&Tool{Config: ToolConfig{configPath}},
	}
	createJSON(conTSptr, "conToolSlicePTR.json", t)

	conTM := map[string]Tool{
		"test1": Tool{Config: ToolConfig{configPath}},
		"test2": Tool{Config: ToolConfig{configPath}},
	}
	createYAML(conTM, "conToolMap.yml", t)
	createTOML(conTM, "conPTRToolMap.toml", t)

	conTMptr := map[string]*Tool{
		"test1": &Tool{Config: ToolConfig{configPath}},
		"test2": &Tool{Config: ToolConfig{configPath}},
	}
	createJSON(conTMptr, "conToolMapPTR.json", t)

	defer removeConfigFiles(t)

	SetFileSearchCaseSensitive(true)

	PrintInfo()

	var test Box
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}

	assert.Equal(t, configPath, test.SubBox.Tool1.Config.Path, "subBox not correctly loaded")
	assert.Equal(t, configPath, test.Tool.Config.Path, "test.Tool.Config.Path is empty")
	assert.Equal(t, configPath, test.PTRTool.Config.Path, "test.PTRTool.Config.Path is empty")
	assert.Equal(t, 0, len(test.ToolNoConfigurable.Path), "test.ToolNoConfigurable.Path:", test.ToolNoConfigurable.Path)
	assert.Equal(t, 0, len(test.PTRToolNoConfigurable.Path), "test.PTRToolNoConfigurable.Path:", test.PTRToolNoConfigurable.Path)

	assert.Equal(t, configPath, test.ToolSlice[0].Config.Path, "test.ToolSlice.Config.Path is empty")
	assert.Equal(t, configPath, test.ToolSlicePTR[0].Config.Path, "test.ToolSlicePTR.Config.Path is empty")
	assert.Equal(t, configPath, (*test.PTRToolSlice)[0].Config.Path, "test.PTRToolSlice.Config.Path is empty")
	assert.Equal(t, configPath, test.ToolMap["test1"].Config.Path, "test.ToolMap.Config.Path is empty")
	assert.Equal(t, configPath, test.ToolMapPTR["test1"].Config.Path, "test.ToolMapPTR.Config.Path is empty")
	assert.Equal(t, configPath, (*test.PTRToolMap)["test1"].Config.Path, "test.PTRToolMap.Config.Path is empty")

	assert.Equal(t, configPath, test.ConSlice[0].Config.Path, "test.ConSlice[0].Config.Path:", test.ConSlice[0].Config.Path)
	assert.Equal(t, configPath, (*test.ConSlicePtr)[0].Config.Path, "(*test.ConSlicePtr)[0].Config.Path:", (*test.ConSlicePtr)[0].Config.Path)
	assert.Equal(t, configPath, test.ConMap["test1"].Config.Path, "test.ConMap['test1'].Config.Path:", test.ConMap["test1"].Config.Path)
	assert.Equal(t, configPath, (*test.ConMapPtr)["test1"].Config.Path, "(*test.ConMapPtr)['test1'].Config.Path:", (*test.ConMapPtr)["test1"].Config.Path)

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
	assert.NotEqual(t, 0, len(test1.Tool1.Config.Path), "test1.Tool1.Config.Path:", test1.Tool1.Config.Path)
	assert.NotEqual(t, 0, len(test1.Tool2.Config.Path), "test1.Tool2.Config.Path:", test1.Tool2.Config.Path)

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
	assert.NotEqual(t, 0, len(test3.Tool1.Config.Path), "test3.Tool1.Config.Path:", test3.Tool1.Config.Path)
	assert.NotEqual(t, 0, len(test3.Tool2.Config.Path), "test3.Tool2.Config.Path:", test3.Tool2.Config.Path)

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
	assert.NotEqual(t, 0, len(test.Tool1.Config.Path), "test.Tool1.Path:", test.Tool1.Config.Path)
	assert.NotEqual(t, 0, len(test.Tool2.Config.Path), "test.Tool2.Path:", test.Tool2.Config.Path)
	assert.NotEqual(t, 0, len(test.Tool3.Config.Path), "test.Tool3.Path:", test.Tool3.Config.Path)
}

func TestConfigFileNotFound(t *testing.T) {
	var test BoxConfigFiles
	assert.Error(t, LoadToolBox(&test, configPath), "unexistent config file does not return error")
}

type BoxTags struct {
	Tool1 Tool
	Tool2 Tool  `sprbox:"-"`
	Tool3 Tool  `sprbox:"test.yml"`
	Tool5 Tool  `sprbox:"-"`
	Tool6 *Tool `sprbox:"-"`
	Tool7 *Tool
	Tool8 *Tool `sprbox:"tool8"`
}

func TestBoxTags(t *testing.T) {
	BUILDENV = "dev"

	devConfig := defaultToolConfig
	devpath := "dev"
	devConfig.Path = devpath

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
	assert.Equal(t, defaultToolConfig.Path, test.Tool1.Config.Path, "test.Tool1.Config.Path:", test.Tool1.Config.Path)
	assert.NotEqual(t, defaultToolConfig.Path, test.Tool2.Config.Path, "test.Tool2.Config.Path:", test.Tool2.Config.Path)
	assert.Equal(t, defaultToolConfig.Path, test.Tool3.Config.Path, "test.Tool3.Config.Path:", test.Tool3.Config.Path)
	assert.Equal(t, 0, len(test.Tool5.Config.Path), "test.Tool5.Config.Path:", test.Tool5.Config.Path)
	assert.NotEqual(t, defaultToolConfig.Path, test.Tool6.Config.Path, "test.Tool6.Config.Path:", test.Tool6.Config.Path)
	assert.Equal(t, devpath, test.Tool7.Config.Path, "test.Tool7.Path:", test.Tool7.Config.Path)
	assert.Equal(t, devpath, test.Tool8.Config.Path, "test.Tool8.Path:", test.Tool8.Config.Path)
}

type BoxAfterConfig struct {
	Tool1 Tool
	Tool2 Tool `sprbox:"-"`
}

func TestBoxAfterConfig(t *testing.T) {
	createYAML(defaultToolConfig, "Tool1.yml", t)
	defer removeConfigFiles(t)

	tString := "must remain the same"
	test := BoxAfterConfig{}
	test.Tool2 = Tool{Config: ToolConfig{Path: tString}}
	if err := LoadToolBox(&test, configPath); err != nil {
		t.Error(err)
	}

	assert.NotEqual(t, 0, len(test.Tool1.Config.Path), "test1.Config.Path:", test.Tool1.Config.Path)
	assert.Equal(t, tString, test.Tool2.Config.Path, "test2.Path:", test.Tool2.Config.Path)
}
