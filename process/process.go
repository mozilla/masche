// this package provides functions to interact with the os processes
// You can list all the processes running on the os, filter them via a regexp
// and then use them from in other masche modules, because they are already open.
package process

import "regexp"

type Process interface {
	Pid() uint
	Name() (name string, harderror error, softerrors []error)
	Close() (harderror error, softerrors []error)
	Handle() interface{} // OS Specific Internal Handle.
}

func OpenFromPid(pid uint) (p Process, harderror error, softerrors []error) {
	return openFromPid(pid)
}

func AllPids() (pids []uint, harderror error, softerrors []error) {
	return allPids()
}

func AllProcesses() (ps []Process, harderror error, softerrors []error) {
	pids, err, _ := AllPids()
	if err != nil {
		return nil, err, nil
	}

	ps = make([]Process, 0)
	softerrs := make([]error, 0)
	for _, pid := range pids {
		p, err, softs := OpenFromPid(pid)
		if err != nil {
			softerrs = append(softerrs, err)
		}
		if softs != nil {
			softerrs = append(softerrs, softs...)
		}
		ps = append(ps, p)
	}
	return ps, nil, softerrs
}

func CloseAll(ps []Process) (harderrors []error, softerrors []error) {
	harderrors = make([]error, 0)
	softerrors = make([]error, 0)

	for _, p := range ps {
		hard, soft := p.Close()
		if hard != nil {
			harderrors = append(harderrors, hard)
		}
		if soft != nil {
			softerrors = append(softerrors, soft...)
		}
	}

	return harderrors, softerrors
}

func Grep(r *regexp.Regexp) (ps []Process, harderror error, softerrors []error) {
	procs, harderror, softerrors := AllProcesses()
	if harderror != nil {
		return nil, harderror, nil
	}

	matchs := make([]Process, 0)

	for _, p := range procs {
		name, err, softs := p.Name()
		if err != nil {
			softerrors = append(softerrors, err)
		}
		if softs != nil {
			softerrors = append(softerrors, softs...)
		}

		if r.MatchString(name) {
			matchs = append(matchs, p)
		} else {
			p.Close()
		}
	}

	return matchs, nil, softerrors
}
