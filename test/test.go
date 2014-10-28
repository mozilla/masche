// This package contains utility methos for testing
package test

import (
	"fmt"
	"os"
	"os/exec"
)

func PrintSoftErrors(softerrors []error) {
	for _, err := range softerrors {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func LaunchTestCase() (*exec.Cmd, error) {
	//TODO: Right now the command is hardcoded. We should decide how to fix this.
	cmd := exec.Command("../test/tools/test.exe")
	err := cmd.Start()
	return cmd, err
}

func LaunchTestCaseAndWaitForInitialization() (*exec.Cmd, error) {
	//TODO: Right now the command is hardcoded. We should decide how to fix this.
	return launchProcessAndWaitInitialization("../test/tools/test.exe")
}

// starts a process and waits until it writes something to stdout: that way we know it has been initialized.
// the process launched should write to stdout once it has been fully initialized.
func launchProcessAndWaitInitialization(file string) (*exec.Cmd, error) {
	cmd := exec.Command(file)

	childout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer childout.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Wait until the process writes something to stdout, so we know it has initialized all its memory.
	if read, err := childout.Read(make([]byte, 1)); err != nil || read != 1 {
		cmd.Process.Kill()
		return nil, err
	}

	return cmd, nil
}
