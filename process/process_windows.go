package process

import "errors"

func (p process) Name() (name string, harderror error, softerrors []error) {
	return "", errors.New("NOT YET IMPLEMENTED"), nil
}

func getAllPids() (pids []uint, harderror error, softerrors []error) {
	return nil, errors.New("NOT YET IMPLEMENTED"), nil
}
