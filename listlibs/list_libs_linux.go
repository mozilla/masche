package listlibs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// HasLibrary checks if a process with a given pid has a certain library.
// To do this we use the /proc/<pid>/maps to know which files are mapped.
// the file format is described in `man proc`
func HasLibrary(pid uint, r *regexp.Regexp) (bool, error) {
	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Just keep the last part of the mapped filename
		// TODO(mvanotti): Probably now that we are using regexp,
		// we may want to do the regexp over the whole filename.
		fields := strings.Split(line, "/")
		if len(fields) <= 1 {
			continue
		}
		library := fields[len(fields)-1]

		if r.MatchString(library) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// Softerror describes an error related to a particular process.
type Softerror struct {
	Pid uint
	Err error
}

func (s Softerror) Error() string {
	return fmt.Sprintf("Pid: %d; Error: %v", s.Pid, s.Err)
}

// FindProcWithLib lists all the process that have loaded a library whose name matches
// the given regexp.
// It works loo	king at all the pids listed in /proc folder, and for each of them, checking its maps file.
// This function returns the list of the process ids of the matching processes.
// There may be some process that couldn't be opened or failed to list their libraries,
// those processes are returned as Softerrors (it means that the rest of the listed processes are OK).
// If there function fails and the results are invalid, a normal error will be returned.
func FindProcWithLib(r *regexp.Regexp) ([]uint, []Softerror, error) {
	files, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, nil, err
	}

	res := make([]uint, 0)
	errs := make([]Softerror, 0)

	for _, f := range files {
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}

		if has, err := HasLibrary(uint(pid), r); err != nil {
			errs = append(errs, Softerror{uint(pid), err})
		} else if has {
			res = append(res, uint(pid))
		}
	}

	return res, errs, nil
}
