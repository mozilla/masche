package listlibs

import (
	"regexp"

	"github.com/mozilla/masche/process"
)

// ListLoadedLibraries lists all the libraries (their absolute paths) loaded by a process.
func ListLoadedLibraries(p process.Process) (libraries []string, softerrors []error, harderror error) {
	return listLoadedLibraries(p)
}

// GetMatchingLoadedLibraries lists the libraries loaded by process p whose path matches r.
func GetMatchingLoadedLibraries(p process.Process, r *regexp.Regexp) (libraries []string, softerrors []error,
	harderror error,
) {

	allLibraries, softerrors, harderror := ListLoadedLibraries(p)
	if harderror != nil {
		return
	}

	for _, lib := range allLibraries {
		if r.MatchString(lib) {
			libraries = append(libraries, lib)
		}
	}

	return
}
