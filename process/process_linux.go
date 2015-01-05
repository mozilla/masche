package process

import (
	"fmt"
	"io/ioutil"
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

func (p proc) Handle() interface{} {
	return nil
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
	return proc(pid), nil, nil
}
