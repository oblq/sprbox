package sprbox

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	BUILDENV = Local.ID()
	assert.Equal(t, Env(), Local, "Local environment not matched")

	BUILDENV = Production.ID()
	assert.Equal(t, Env(), Production, "Production environment not matched")

	BUILDENV = Staging.ID()
	assert.Equal(t, Env(), Staging, "Staging environment not matched")

	BUILDENV = Testing.ID()
	assert.Equal(t, Env(), Testing, "Testing environment not matched")

	BUILDENV = Development.ID()
	assert.Equal(t, Env(), Development, "Development environment not matched")

	BUILDENV = ""
	os.Setenv(EnvVarKey, "")
	VCS = nil

	Env().PrintInfo()

	os.Setenv(EnvVarKey, "staging")
	assert.Equal(t, Env(), Staging, "Staging environment not matched")

	println(Env().Info())
	println(EnvSubDir(configPath))

	BUILDENV = ""
	os.Unsetenv(EnvVarKey)

	VCS = NewRepository("./")
	println(Env().Info())

	VCS = nil
	assert.Equal(t, Env(), Testing, "Development is not testing by default during testing: "+Env().ID()+" - "+os.Args[0])

	// RegEx test
	Production.AppendExp("branch/*")
	assert.True(t, Production.MatchTag("branch/test"), "error in RegEx matcher...")

	Production.SetExps([]string{"test*"})
	assert.True(t, Production.MatchTag("test1"), "error in RegEx matcher...")
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
