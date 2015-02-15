package process

import (
	"log"
	"regexp"
	"testing"

	"github.com/mozilla/masche/test"
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

// TODO(mvanotti): Fix this test for windows:
// Right now windows is reporting names in Device Format, which is different to the TestCasePath
func testProcessName(t *testing.T) {
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
		//t.Error("Expected name", test.GetTestCasePath(), "and got", name)
	}
}

// TODO(mvanotti): Fix this test for windows:
// Right now windows is reporting names in Device Format, which is different to the TestCasePath
func TestOpenByName(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	r := regexp.MustCompile("test[/\\\\]tools[/\\\\]test")
	procs, err, softerrors := OpenByName(r)
	defer CloseAll(procs)
	test.PrintSoftErrors(softerrors)
	if len(procs) == 0 {
		t.Error("The test case was launched and not opened.")
	}

	for _, proc := range procs {
		name, err, softerrors := proc.Name()
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}

		if name != test.GetTestCasePath() {
			//t.Error("Expected name", test.GetTestCasePath(), "and got", name)
		}
	}
}
