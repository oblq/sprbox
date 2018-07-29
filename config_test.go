package sprbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

const configPath = "/tmp/sprbox"

func createYAML(object interface{}, fileName string, t *testing.T) {
	bytes, err := yaml.Marshal(object)
	if err != nil {
		t.Errorf("failed to create config file: %v", err)
	}
	writeFiles(fileName, bytes, t)
}

func createTOML(object interface{}, fileName string, t *testing.T) {
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(object); err != nil {
		t.Errorf("failed to create config file: %v", err)
	}
	writeFiles(fileName, buffer.Bytes(), t)
}

func createJSON(object interface{}, fileName string, t *testing.T) {
	bytes, err := json.Marshal(object)
	if err != nil {
		t.Errorf("failed to create config file: %v", err)
	}
	writeFiles(fileName, bytes, t)
}

func writeFiles(fileName string, bytes []byte, t *testing.T) {
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		t.Error(err)
	}

	filePath := filepath.Join(configPath, fileName)
	if err := ioutil.WriteFile(filePath, bytes, os.ModePerm); err != nil {
		t.Errorf("failed to create config file: %v", err)
	}

	fileNameExt := filepath.Ext(fileName)
	fileNameNoExt := strings.TrimSuffix(fileName, fileNameExt)
	fileNameEnv := fmt.Sprintf("%s.%s%s", fileNameNoExt, Env().String(), fileNameExt)
	fileEnvPath := filepath.Join(configPath, fileNameEnv)
	if err := ioutil.WriteFile(fileEnvPath, bytes, os.ModePerm); err != nil {
		t.Errorf("failed to create config file: %v", err)
	}
}

func removeConfigFiles(t *testing.T) {
	if err := os.RemoveAll(configPath); err != nil {
		t.Error(err)
	}
}

type Postgres struct {
	DB       string `sprbox:"env=POSTGRES_DB,default=postgres"`
	User     string `sprbox:"env=POSTGRES_USER,default=postgres"`
	Password string `sprbox:"env=POSTGRES_PASSWORD,required"`
	Port     int    `sprbox:"default=5432"`
}

type EmbeddedStruct struct {
	Field1 string `sprbox:"default=sprbox"`
	Field2 string `sprbox:"required"`
}

type Config struct {
	String        string `sprbox:"default=sprbox"`
	PG            Postgres
	Slice         []string
	Map           *map[string]string
	EmbeddedSlice []EmbeddedStruct
	// EmbeddedStruct without pointer inside of a map would not be addressable,
	// so, this is the way that make sense...
	// Otherwise also 'config.EmbeddedMap["test"].Field1 = "a value"' can't be done.
	EmbeddedMap map[string]*EmbeddedStruct
}

func defaultConfig() Config {
	config := Config{
		String: "sprbox",
		Slice:  []string{"elem1", "elem2"},
		Map:    &map[string]string{"key": "value"},
		PG: Postgres{
			DB:       "sprbox",
			User:     "me",
			Password: "myPass123",
			Port:     5432,
		},
		EmbeddedSlice: []EmbeddedStruct{
			{
				Field1: "sprbox",
				Field2: "f2",
			},
		},
		EmbeddedMap: map[string]*EmbeddedStruct{
			"test": {
				Field1: "sprbox",
				Field2: "f2map",
			},
		},
	}
	return config
}

func TestYAML(t *testing.T) {
	config := defaultConfig()
	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	Debug()

	var result1 Config
	if err := LoadConfig(&result1, filepath.Join(configPath, fileName)); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result1, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result1)
	}

	var result2 Config
	if err := LoadConfig(&result2, filepath.Join(configPath, "config")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result2, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result2)
	}

	debug = false
}

func TestYML(t *testing.T) {
	config := defaultConfig()
	fileName := "config.yml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result1 Config
	if err := LoadConfig(&result1, filepath.Join(configPath, fileName)); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result1, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result1)
	}

	var result2 Config
	if err := LoadConfig(&result2, filepath.Join(configPath, "config")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result2, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result2)
	}
}

func TestTOML(t *testing.T) {
	config := defaultConfig()
	fileName := "config.toml"
	createTOML(config, fileName, t)
	defer removeConfigFiles(t)

	var result1 Config
	if err := LoadConfig(&result1, filepath.Join(configPath, fileName)); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result1, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result1)
	}

	var result2 Config
	if err := LoadConfig(&result2, filepath.Join(configPath, "config")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result2, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result2)
	}
}

func TestJSON(t *testing.T) {
	config := defaultConfig()
	fileName := "config.json"
	createJSON(config, fileName, t)
	defer removeConfigFiles(t)

	var result1 Config
	if err := LoadConfig(&result1, filepath.Join(configPath, fileName)); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result1, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result1)
	}

	var result2 Config
	if err := LoadConfig(&result2, filepath.Join(configPath, "config")); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result2, config) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", config, result2)
	}
}

// only passing filename
func TestYAMLWrongPath(t *testing.T) {
	fileName := "config.yaml"
	var result1 Config
	if err := LoadConfig(&result1, fileName); err == nil {
		t.Error(err)
	}
}

//SFT = struct field tags
func TestSFTDefault(t *testing.T) {
	config := defaultConfig()
	config.String = ""
	config.PG.Port = 0
	config.EmbeddedSlice[0].Field1 = ""
	config.EmbeddedMap["test"].Field1 = ""

	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	Debug()

	var result Config
	LoadConfig(&result, filepath.Join(configPath, fileName))
	if !reflect.DeepEqual(result, defaultConfig()) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", defaultConfig(), result)
	}

	debug = false
}

//SFT = struct field tags
func TestSFTRequired(t *testing.T) {
	config := defaultConfig()
	config.PG.Password = ""

	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result Config
	if err := LoadConfig(&result, filepath.Join(configPath, fileName)); err == nil {
		t.Errorf("should return error if a required field is missing ")
	}
}

//SFT = struct field tags
func TestSFTEnv(t *testing.T) {
	config := defaultConfig()
	config.PG.DB = "wrong"
	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result Config
	os.Setenv("POSTGRES_DB", "postgres")
	defer os.Unsetenv("POSTGRES_DB")
	LoadConfig(&result, filepath.Join(configPath, fileName))
	if result.PG.DB != "postgres" {
		t.Errorf("env var not loaded correctly")
	}
}

func TestCorruptedFile(t *testing.T) {
	fileName := "config.yaml"
	createYAML("wrongObject", fileName, t)
	defer removeConfigFiles(t)

	var result Config
	if err := LoadConfig(&result, filepath.Join(configPath, fileName)); err == nil {
		t.Errorf("corrupted file does not return error")
	}
}

func TestWrongConfigFileName(t *testing.T) {
	config := defaultConfig()
	fileName := "config.wrong"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result Config
	if err := LoadConfig(&result, filepath.Join(configPath, fileName)); err == nil {
		t.Errorf("wrong path does not return error")
	}
}

func TestNotAStruct(t *testing.T) {
	config := defaultConfig()
	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result string
	if err := LoadConfig(&result, filepath.Join(configPath, fileName)); err == nil {
		t.Error(err)
	}
}

func TestNoFileName(t *testing.T) {
	config := defaultConfig()
	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result1 Config
	if err := LoadConfig(&result1, configPath); err == nil {
		t.Error(err)
	}
}

func TestEnvironmentFiles(t *testing.T) {
	BUILDENV = "dev"

	config := Tool{}
	createYAML(config, "tool1.yml", t)
	createJSON(config, "tool."+Env().String()+".json", t)
	createTOML(config, "tool.toml", t)
	defer removeConfigFiles(t)

	// '<path>/<file>.<environment>.*'
	if files := configFilesByEnv(filepath.Join(configPath, "tool")); len(files) == 1 {
		if files[0] != filepath.Join(configPath, "tool."+Env().String()+".json") {
			t.Error("file not matched")
		}
	}

	// '<path>/<file>.*'
	if files := configFilesByEnv(filepath.Join(configPath, "tool1")); len(files) == 1 {
		if files[0] != filepath.Join(configPath, "tool1.yaml") {
			t.Error("file not matched")
		}
	}

	// '<path>/<file>.<ext>'
	if files := configFilesByEnv(filepath.Join(configPath, "tool.toml")); len(files) == 1 {
		if files[0] != filepath.Join(configPath, "tool.toml") {
			t.Error("file not matched")
		}
	}

	// wrong ext '<path>/<file>.<ext>'
	if files := configFilesByEnv(filepath.Join(configPath, "tool2.toml")); len(files) > 1 {
		t.Error("file not matched")
	}

	// case insensitive '<path>/<file>.<environment>.*'
	FileSearchCaseSensitive = false
	if files := configFilesByEnv(filepath.Join(configPath, "TOOL")); len(files) == 1 {
		if files[0] != filepath.Join(configPath, "tool."+Env().String()+".json") {
			t.Error("file not matched")
		}
	}
}
