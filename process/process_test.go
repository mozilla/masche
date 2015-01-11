package process

import (
	"github.com/mozilla/masche/test"
	"testing"
)

func TestOpenFromPid(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	proc, err, softerrors := OpenFromPid(pid)
	defer proc.Close()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcessName(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	proc, err, softerrors := OpenFromPid(pid)
	defer proc.Close()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	name, err, softerrors := proc.Name()
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if name != test.GetTestCasePath() {
		t.Error("Expected name", test.GetTestCasePath(), "and got", name)
	}
}
