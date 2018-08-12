package sprbox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// files type regexp
const regexExt = `(?i)(.y(|a)ml|.toml|.json)` // `(?i)(\..{3,4})` //

// FILE SEARCH ---------------------------------------------------------------------------------------------------------

// walkConfigPath look for a file matching the passed regex skipping sub-directories.
func walkConfigPath(configPath string, regex *regexp.Regexp) (matchedFile string) {
	err := filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		// nil if the path does not exist
		if info == nil {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() != filepath.Base(configPath) {
			return filepath.SkipDir
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if regex.MatchString(info.Name()) {
			matchedFile = path
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	return
}

// configFilesByEnv will search for the given file in the given path
// returning all the eligible files (eg.: <path>/myConfig.yaml and <path>/myConfig.<environment>.yaml)
//
// 'filePath' can also be passed without file extension,
// searchFilesByEnv is agnostic and will match any
// supported extension in that case.
//
// The 'file' name will be searched as (in that order):
//  - '<path>/<file>(.* || <the_provided_extension>)'
//  - '<path>/<file>.<environment>(.* || <the_provided_extension>)'
//
// The latest found files will override previous.
func configFilesByEnv(files ...string) (foundFiles []string) {
	for _, file := range files {
		configPath, fileName := filepath.Split(file)
		if len(configPath) == 0 {
			configPath = "./"
		}

		var regexEnv *regexp.Regexp
		var regex *regexp.Regexp

		ext := filepath.Ext(fileName)
		extTrimmed := strings.TrimSuffix(fileName, ext)
		if len(ext) == 0 {
			ext = regexExt
			debugPrintf(darkGrey("\nlooking for '%s%s' in '%s'..."), fileName, regexExt, configPath)
		} else {
			debugPrintf(darkGrey("\nlooking for '%s' in '%s'..."), fileName, configPath)
		}

		format := "^%s%s$"
		if !fileSearchCaseSensitive {
			format = "(?i)(^%s)%s$"
		}
		regexEnv = regexp.MustCompile(fmt.Sprintf(format, fmt.Sprintf("%s.%s", extTrimmed, Env().ID()), ext))
		regex = regexp.MustCompile(fmt.Sprintf(format, extTrimmed, ext))

		// look for the config file in the config path (eg.: tool.yml)
		if matchedFiles := walkConfigPath(configPath, regex); len(matchedFiles) > 0 {
			foundFiles = append(foundFiles, matchedFiles)
		}

		// look for the env config file in the config path (eg.: tool.development.yml)
		if matchedFiles := walkConfigPath(configPath, regexEnv); len(matchedFiles) > 0 {
			foundFiles = append(foundFiles, matchedFiles)
		}
	}

	if len(foundFiles) > 0 {
		debugPrintf("\n%s %s", strings.Join(foundFiles, green(" <- ")), green("="))
	} else {
		debugPrintf("\n")
	}
	return
}
