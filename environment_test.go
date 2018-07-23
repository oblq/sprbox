package sprbox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironment(t *testing.T) {
	BUILDENV = Local.String()
	if Env() != Local {
		t.Error("Local environment not matched")
	}

	BUILDENV = Production.String()
	if Env() != Production {
		t.Error("Production environment not matched")
	}

	BUILDENV = Staging.String()
	if Env() != Staging {
		t.Error("Staging environment not matched")
	}

	BUILDENV = Testing.String()
	if Env() != Testing {
		t.Error("Testing environment not matched")
	}

	BUILDENV = Development.String()
	if Env() != Development {
		t.Error("Development environment not matched")
	}

	BUILDENV = ""
	os.Setenv(EnvVarKey, "")
	VCS = nil

	Env().PrintInfo()

	os.Setenv(EnvVarKey, "staging")
	if Env() != Staging {
		t.Error("Staging environment not matched")
	}

	println(Env().Info())

	// RegEx test
	Production.AppendExp("branch/*")
	if !Production.MatchTag("branch/test") {
		t.Error("error in RegEx matcher...")
	}

	Production.SetExps([]string{"test*"})
	if !Production.MatchTag("test1") {
		t.Error("error in RegEx matcher...")
	}
}

const configPath = "/tmp/sprbox"

func CreateConfigFiles(subPath string, namelist []string, t *testing.T) {
	envPath := filepath.Join(configPath, subPath)
	if err := os.MkdirAll(envPath, os.ModePerm); err != nil {
		t.Error(err)
	}
	for _, name := range namelist {
		_, err := os.Create(filepath.Join(envPath, name))
		if err != nil {
			t.Error(err)
		}
	}
}

func RemoveConfigFiles(env string, t *testing.T) {
	envPath := filepath.Join(configPath, env)
	if err := os.RemoveAll(envPath); err != nil {
		t.Error(err)
	}
}

func TestEnvironmentPath(t *testing.T) {
	BUILDENV = "dev"

	CreateConfigFiles("", []string{"tool." + Env().String() + ".json", "tool.toml", "tool1.yaml"}, t)
	defer RemoveConfigFiles("", t)

	CreateConfigFiles(Env().String(), []string{"tool2.json"}, t)
	defer RemoveConfigFiles(Env().String(), t)

	// '<path>/<file>.<environment>.*'
	filePath := SearchFileByEnv(configPath, "tool", true)
	if filePath != filepath.Join(configPath, "tool."+Env().String()+".json") {
		t.Error("file not matched")
	}

	// '<path>/<file>.*'
	filePath = SearchFileByEnv(configPath, "tool1", true)
	if filePath != filepath.Join(configPath, "tool1.yaml") {
		t.Error("file not matched")
	}

	// '<path>/<environment>/<file>.*
	filePath = SearchFileByEnv(configPath, "tool2", true)
	if filePath != filepath.Join(configPath, Env().String(), "tool2.json") {
		t.Error("file not matched")
	}

	// '<path>/<file>.<ext>'
	filePath = SearchFileByEnv(configPath, "tool.toml", true)
	if filePath != filepath.Join(configPath, "tool.toml") {
		t.Error("file not matched")
	}

	// wrong ext '<path>/<file>.<ext>'
	filePath = SearchFileByEnv(configPath, "tool2.toml", true)
	if len(filePath) > 0 {
		t.Error("file not matched")
	}

	// case insensitive '<path>/<file>.<environment>.*'
	filePath = SearchFileByEnv(configPath, "TOOL", false)
	if filePath != filepath.Join(configPath, "tool."+Env().String()+".json") {
		t.Error("file not matched")
	}
}
