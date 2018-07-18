package sprbox

import (
	"errors"
	"path/filepath"
	"testing"
)

// add	debug = true for details

// TestBox is a struct implementing 'autobox' interface.
type Tool struct {
	ConfigPath string
}

// Go2Box is the 'autobox' interface implementation.
func (c *Tool) Go2Box(configPath string) error {
	c.ConfigPath = configPath
	return nil
}

// TestBoxError is a struct implementing 'autobox' interface.
type ToolError struct {
	ConfigPath string
}

// Go2Box is the 'autobox' interface implementation.
func (c *ToolError) Go2Box(configPath string) error {
	return errors.New("fake error for test")
}

// TestBoxError is a struct implementing 'autobox' interface.
type ToolNoAutobox struct {
	ConfigPath string
}

type Box struct {
	Tool             Tool
	PTRTool          *Tool
	ToolNoAutobox    ToolNoAutobox
	PTRToolNoAutobox *ToolNoAutobox
}

func TestBox(t *testing.T) {
	PrintInfo(false)
	var test Box
	if err := InitAndConfig(&test, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if len(test.Tool.ConfigPath) == 0 {
		t.Error("test.Tool.ConfigPath is empty")
	}
	if len(test.PTRTool.ConfigPath) == 0 {
		t.Error("test.PTRTool.ConfigPath is empty")
	}
	if len(test.ToolNoAutobox.ConfigPath) > 0 {
		t.Error("test.ToolNoAutobox.ConfigPath:", test.ToolNoAutobox.ConfigPath)
	}
	if len(test.PTRToolNoAutobox.ConfigPath) > 0 {
		t.Error("test.PTRToolNoAutobox.ConfigPath:", test.PTRToolNoAutobox.ConfigPath)
	}
}

type BoxError struct {
	ToolError ToolError
}

func TestBoxError(t *testing.T) {
	var test BoxError
	if err := InitAndConfig(&test, "testConfigPath"); err == nil {
		t.Error(err)
	}
}

type PTRToolError struct {
	PTRToolError *ToolError
}

func TestPTRToolError(t *testing.T) {
	var test PTRToolError
	if err := InitAndConfig(&test, "testConfigPath"); err == nil {
		t.Error(err)
	}
}

type BoxNil struct {
	Tool1 Tool
	Tool2 *Tool
}

func TestNilBox(t *testing.T) {
	ColoredLog = false
	var test1 BoxNil
	if err := InitAndConfig(&test1, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if len(test1.Tool1.ConfigPath) == 0 {
		t.Error("test1.Tool1.ConfigPath:", test1.Tool1.ConfigPath)
	}
	if len(test1.Tool2.ConfigPath) == 0 {
		t.Error("test1.Tool2.ConfigPath:", test1.Tool2.ConfigPath)
	}

	var test2 *BoxNil
	if err := InitAndConfig(test2, "testConfigPath"); err != nil {
		t.Log(err)
	} else {
		t.Error(err)
	}

	var test3 = &BoxNil{}
	if err := InitAndConfig(test3, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if len(test3.Tool1.ConfigPath) == 0 {
		t.Error("test3.Tool1.ConfigPath:", test3.Tool1.ConfigPath)
	}
	if len(test3.Tool2.ConfigPath) == 0 {
		t.Error("test3.Tool2.ConfigPath:", test3.Tool2.ConfigPath)
	}
}

func TestToolItself(t *testing.T) {
	var test Tool
	if err := InitAndConfig(&test, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if len(test.ConfigPath) == 0 {
		t.Error("test.ConfigPath:", test.ConfigPath)
	}
}

type BoxTags struct {
	Tool1 Tool
	Tool2 Tool  `sprbox:"omit"`
	Tool3 Tool  `sprbox:"test.yml"`
	Tool4 Tool  `sprbox:"omit,test.yml"`
	Tool5 Tool  `sprbox:"test.yml,omit"`
	Tool6 *Tool `sprbox:"omit"`
}

func TestBoxTags(t *testing.T) {
	var test BoxTags
	if err := InitAndConfig(&test, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if filepath.Base(test.Tool1.ConfigPath) != "Tool1.yml" {
		t.Error("test.Tool1.ConfigPath:", test.Tool1.ConfigPath)
	}
	if len(test.Tool2.ConfigPath) > 0 {
		t.Error("test.Tool2.ConfigPath:", test.Tool2.ConfigPath)
	}
	if filepath.Base(test.Tool3.ConfigPath) != "test.yml" {
		t.Error("test.Tool3.ConfigPath:", test.Tool3.ConfigPath)
	}
	if len(test.Tool4.ConfigPath) > 0 {
		t.Error("test.Tool4.ConfigPath:", test.Tool4.ConfigPath)
	}
	if len(test.Tool5.ConfigPath) > 0 {
		t.Error("test.Tool5.ConfigPath:", test.Tool5.ConfigPath)
	}
	if test.Tool6 != nil {
		t.Error("test.Tool6 not nil", test.Tool6)
	}
}

type BoxAfterConfig struct {
	Tool1 Tool
	Tool2 Tool `sprbox:"omit"`
}

func TestBoxAfterConfig(t *testing.T) {
	tString := "must remain the same"
	test := BoxAfterConfig{}
	test.Tool2 = Tool{ConfigPath: tString}
	if err := InitAndConfig(&test, "testConfigPath"); err != nil {
		t.Error(err)
	}
	if len(test.Tool1.ConfigPath) == 0 {
		t.Error("test1.ConfigPath:", test.Tool1.ConfigPath)
	}
	if test.Tool2.ConfigPath != tString {
		t.Error("test2.ConfigPath:", test.Tool2.ConfigPath)
	}
}

func TestNotAStructErr(t *testing.T) {
	test := []string{"test"}
	if err := InitAndConfig(&test, "testConfigPath"); err != errNotAStructPointer {
		t.Error(err)
	}
}

// TestBoxError is a struct implementing 'autobox' interface.
type ToolNotAStruct []string

// Go2Box is the 'autobox' interface implementation.
func (c *ToolNotAStruct) Go2Box(configPath string) error {
	return nil
}

type BoxNotAStructTool struct {
	NotAStruct ToolNotAStruct
}

func TestToolNotAStruct(t *testing.T) {
	debug = true
	var test BoxNotAStructTool
	if err := InitAndConfig(&test, "testConfigPath"); err != nil {
		t.Error(err)
	}
}
