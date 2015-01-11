package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type proc uint

func (p proc) Pid() uint {
	return uint(p)
}

func (p proc) Name() (name string, harderror error, softerrors []error) {
	exename := filepath.Join("/proc", fmt.Sprintf("%d", p.Pid()), "exe")
	name, err := filepath.EvalSymlinks(exename)

	return name, err, nil
}

func (p proc) Close() (harderror error, softerrors []error) {
	return nil, nil
}

func (p proc) Handle() uintptr {
	return uintptr(p)
}

func getAllPids() (pids []uint, harderror error, softerrors []error) {
	files, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, err, nil
	}

	pids = make([]uint, 0)

	for _, f := range files {
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}
		pids = append(pids, uint(pid))
	}

	return pids, nil, nil
}

func openFromPid(pid uint) (p Process, harderror error, softerrors []error) {
	// Check if we have premissions to read the process memory
	memPath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
	memFile, err := os.Open(memPath)
	if err != nil {
		harderror = fmt.Errorf("Permission denied to access memory of process %v", pid)
		return
	}
	defer memFile.Close()

	return proc(pid), nil, nil
}
