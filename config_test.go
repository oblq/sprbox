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

	SetDebug(true)

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

	SetDebug(false)
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

func TestParsingIntoNonStruct(t *testing.T) {
	config := defaultConfig()
	fileName := "config.yaml"
	createYAML(config, fileName, t)
	defer removeConfigFiles(t)

	var result1 string
	err := LoadConfig(&result1, filepath.Join(configPath, fileName))
	if err == nil {
		t.Error(err)
	}
	t.Log(err)
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

	SetDebug(true)

	var result Config
	LoadConfig(&result, filepath.Join(configPath, fileName))
	if !reflect.DeepEqual(result, defaultConfig()) {
		t.Errorf("\n\nFile:\n%#v\n\nConfig:\n%#v\n\n", defaultConfig(), result)
	}

	SetDebug(false)
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

func TestMapYAML(t *testing.T) {
	config := defaultConfig()
	createYAML(config, "config1.yaml", t)
	config.String = "overriden1"
	createYAML(config, "config2.yaml", t)
	config.PG.DB = "overriden2"
	createYAML(config, "config3.yaml", t)
	defer removeConfigFiles(t)

	SetDebug(true)

	configMap, err := LoadConfigMap(
		filepath.Join(configPath, "config1.yaml"),
		filepath.Join(configPath, "config2.yaml"),
		filepath.Join(configPath, "config3.yaml"),
	)
	if err != nil {
		t.Error(err)
	}

	dump(configMap)

	if configMap["string"] != "overriden1" {
		t.Error("value not overriden")
	}

	if configMap["pg"].(map[interface{}]interface{})["db"] != "overriden2" {
		t.Error("value not overriden")
	}

	SetDebug(false)
}

func TestMapJSON(t *testing.T) {
	config := defaultConfig()
	createJSON(config, "config1.json", t)
	config.String = "overriden1"
	createJSON(config, "config2.json", t)
	config.PG.DB = "overriden2"
	createJSON(config, "config3.json", t)
	defer removeConfigFiles(t)

	SetDebug(true)

	configMap, err := LoadConfigMap(
		filepath.Join(configPath, "config1.json"),
		filepath.Join(configPath, "config2.json"),
		filepath.Join(configPath, "config3.json"),
	)
	if err != nil {
		t.Error(err)
	}

	if configMap["String"] != "overriden1" {
		t.Error("value not overriden")
	}

	if configMap["PG"].(map[string]interface{})["DB"] != "overriden2" {
		t.Error("value not overriden")
	}

	SetDebug(false)
}

func TestMapTOML(t *testing.T) {
	config := defaultConfig()
	createTOML(config, "config1.toml", t)
	config.String = "overriden1"
	createTOML(config, "config2.toml", t)
	config.PG.DB = "overriden2"
	createTOML(config, "config3.toml", t)
	defer removeConfigFiles(t)

	SetDebug(true)

	configMap, err := LoadConfigMap(
		filepath.Join(configPath, "config1.toml"),
		filepath.Join(configPath, "config2.toml"),
		filepath.Join(configPath, "config3.toml"),
	)
	if err != nil {
		t.Error(err)
	}

	if configMap["String"] != "overriden1" {
		t.Error("value not overriden")
	}

	if configMap["PG"].(map[string]interface{})["DB"] != "overriden2" {
		t.Error("value not overriden")
	}

	SetDebug(false)
}

func TestMapMixed(t *testing.T) {
	config := defaultConfig()
	config.PG.DB = "overridenyml"
	createYAML(config, "config1.yml", t)
	config.String = "overriden1"
	createTOML(config, "config2.toml", t)
	config.PG.DB = "overriden2"
	createJSON(config, "config3.json", t)
	defer removeConfigFiles(t)

	SetDebug(true)

	configMap, err := LoadConfigMap(
		filepath.Join(configPath, "config1.yml"),
		filepath.Join(configPath, "config2.toml"),
		filepath.Join(configPath, "config3.json"),
	)
	if err != nil {
		t.Error(err)
	}

	//fmt.Printf("\ndump: %v\n", configMap)

	if configMap["string"] != "sprbox" {
		t.Error("value not overriden")
	}

	if configMap["pg"].(map[interface{}]interface{})["db"] != "overridenyml" {
		t.Error("value not overriden")
	}

	if configMap["String"] != "overriden1" {
		t.Error("value not overriden")
	}

	if configMap["PG"].(map[string]interface{})["DB"] != "overriden2" {
		t.Error("value not overriden")
	}

	SetDebug(false)
}

func TestMapNoFiles(t *testing.T) {
	_, err := LoadConfigMap(filepath.Join(configPath, "config.yml"))
	if err != nil {
		t.Log(err)
	} else {
		t.Error("unexistent file does not return error")
	}
}
