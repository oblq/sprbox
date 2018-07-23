package sprbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SubPathByEnv returns <path>/<environment>
func SubPathByEnv(path string) string {
	return filepath.Join(path, Env().String())
}

// SearchFileByEnv will search for the given file in the giver path.
// 'file' can also be passed without extension,
// SearchFileByEnv is agnostic and will match eny extension in that case.
// The 'file' name will be searched as (in that order):
// - '<path>/<file>.<environment>(.* || <the_provided_extension>)'
// - '<path>/<file>(.* || <the_provided_extension>)'
// - '<path>/<environment>/<file>(.* || <the_provided_extension>)'
func SearchFileByEnv(path, file string, caseSensitive bool) string {
	var regexEnv *regexp.Regexp
	var regex *regexp.Regexp

	format := "^%s"
	if !caseSensitive {
		format = "(?i)(^%s)"
	}

	ext := filepath.Ext(file)
	if len(ext) == 0 {
		fileNameEnv := fmt.Sprintf("%s.%s", file, Env().String())
		regexEnv = regexp.MustCompile(fmt.Sprintf(format+".*", fileNameEnv))
		regex = regexp.MustCompile(fmt.Sprintf(format+".*", file))
	} else {
		fileNameEnv := fmt.Sprintf("%s.%s%s", strings.TrimSuffix(file, ext), Env().String(), ext)
		regexEnv = regexp.MustCompile(fmt.Sprintf(format, fileNameEnv))
		regex = regexp.MustCompile(fmt.Sprintf(format, file))
	}

	// look for the env config file directly in the config path (eg.: tool.development.yml)
	if matchedFile := walk(path, regexEnv); len(matchedFile) > 0 {
		return matchedFile
	}

	// look for the config file directly in the config path (eg.: tool.yml)
	if matchedFile := walk(path, regex); len(matchedFile) > 0 {
		return matchedFile
	}

	// look for the config file in the config path env sub-directory (eg.: development/tool.yml)
	if matchedFile := walk(SubPathByEnv(path), regex); len(matchedFile) > 0 {
		return matchedFile
	}

	return ""
}

// walk look for a file matching the passed regex skipping sub-directories.
func walk(searchPath string, regex *regexp.Regexp) (matchedFile string) {
	if err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		if info.IsDir() && info.Name() != filepath.Base(searchPath) {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if regex.MatchString(info.Name()) {
			//fmt.Printf("regex '%s' matches '%s'\n", regex.String(), path)
			matchedFile = path
		}

		return nil
	}); err != nil {
		fmt.Printf("error in lookForConfigFile: %s", err.Error())
	}
	return
}
