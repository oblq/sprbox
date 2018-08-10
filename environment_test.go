package sprbox

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	BUILDENV = Local.ID()
	if Env() != Local {
		t.Error("Local environment not matched")
	}

	BUILDENV = Production.ID()
	if Env() != Production {
		t.Error("Production environment not matched")
	}

	BUILDENV = Staging.ID()
	if Env() != Staging {
		t.Error("Staging environment not matched")
	}

	BUILDENV = Testing.ID()
	if Env() != Testing {
		t.Error("Testing environment not matched")
	}

	BUILDENV = Development.ID()
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
	println(EnvSubDir(configPath))

	BUILDENV = ""
	os.Unsetenv(EnvVarKey)

	VCS = NewRepository("./")
	println(Env().Info())

	VCS = nil
	if Env() != Testing {
		Env().PrintInfo()
		t.Error("Development is not testing by default during testing: " + Env().ID() + " - " + os.Args[0])
	}

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

func TestCompiledPath(t *testing.T) {
	BUILDENV = Local.ID()
	assert.Equal(t,
		"../static_files/config",
		CompiledPath("../static_files/config"),
		"compiledPath in non RunCompiled environments is wrong")

	Local.RunCompiled = true
	assert.Equal(t,
		"config",
		CompiledPath("../static_files/config"),
		"compiledPath in RunCompiled environments is wrong")
}
