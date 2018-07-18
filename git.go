package sprbox

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var gitLog = kvLogger{ValuePainter: magenta}

// Repository represent a git repository
type Repository struct {
	Path                           string
	BranchName, Commit, Build, Tag string
	Error                          error
}

// NewRepository return a new Repository instance for the given path
func NewRepository(path string) *Repository {
	repo := &Repository{Path: path}
	repo.UpdateInfo()
	return repo
}

// UpdateInfo grab git info and set 'Error' var eventually.
func (r *Repository) UpdateInfo() {
	r.BranchName, r.Error = r.git("rev-parse", "--abbrev-ref", "HEAD")
	r.Commit, _ = r.git("rev-parse", "--short", "HEAD")
	r.Build, _ = r.git("rev-list", "--all", "--count")
	r.Tag, _ = r.git("describe", "--abbrev=0", "--tags", "--always")
}

// Git is the bash git command.
func (r *Repository) git(params ...string) (string, error) {
	gitBinPath, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("[git] git not installed...")
		return "git not installed...", err
	}

	cmd := exec.Command(gitBinPath, params...)

	if len(r.Path) > 0 {
		cmd.Dir = r.Path
	}

	output, err := cmd.Output()
	if err != nil {
		var gitErrString string
		if exitError, ok := err.(*exec.ExitError); ok {
			gitErrString = string(exitError.Stderr)
		} else {
			gitErrString = err.Error()
		}
		gitErrString = strings.TrimPrefix(gitErrString, "fatal: ")
		gitErrString = strings.TrimSuffix(gitErrString, "\n")
		gitErrString = strings.TrimSuffix(gitErrString, ": .git")
		return gitErrString, errors.New(gitErrString)
	}

	out := strings.TrimSuffix(string(output), "\n")
	if len(out) == 0 {
		out = "-"
	}
	return out, nil
}

// PrintInfo print git data in console.
func (r *Repository) PrintInfo() {
	if r == nil {
		return
	}
	gitLog.Println("Git Branch:", r.BranchName)
	gitLog.Println("Git Commit:", r.Commit)
	gitLog.Println("Git Build:", r.Build)
	gitLog.Println("Git Tag:", r.Tag)
	fmt.Println("")
}
