package masche

import (


)



/*
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

// FindProcWithLib returns a list of process ids that have the given library loaded in memory.
// It works looking at all the pids listed in /proc folder, and for each of them, checking its maps file.
func FindProcWithLib(r *regexp.Regexp) ([]uint, error) {
    files, _ := ioutil.ReadDir("/proc/")
    res := make([]uint, 0)

    for _, f := range files {
        pid, err := strconv.Atoi(f.Name())
        if err != nil {
            continue
        }

        if has, err := HasLibrary(uint(pid), r); err != nil {
            //TODO(mvanotti): How should we report errors for multiple files? maybe a map[filepath]error ?
            log.Println(err)
            continue
        } else if has {
            res = append(res, uint(pid))
        }
    }

    return res, nil
}
*/