package sprbox

import "testing"

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

	// RegEx test
	Production.AppendExp("branch/*")
	if !Production.MatchTag("branch/test") {
		t.Error("error in RegEx matcher...")
	}
}
