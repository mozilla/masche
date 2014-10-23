//TODO: Test CopyMemory and WalkMemory. Specially WalkMemory with different buffer sizes, check that they don't
// overlap, and it doesn't skips memory chunks.
package memaccess

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func printSoftErrors(softerrors []error) {
	for _, err := range softerrors {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func TestNewProcessMemoryReader(t *testing.T) {
	//TODO(mvanotti): Right now the command is hardcoded. We should decide how to fix this.
	cmd := exec.Command("../test/tools/test.exe")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	reader, err, softerrors := NewProcessMemoryReader(pid)
	printSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
}

//TODO: Improve this test.
func TestManuallyWalk(t *testing.T) {
	//TODO(mvanotti): Right now the command is hardcoded. We should decide how to fix this.
	cmd := exec.Command("../test/tools/test.exe")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	reader, err, softerrors := NewProcessMemoryReader(pid)
	printSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	var region MemoryRegion
	region, err, _ = reader.NextReadableMemoryRegion(0)
	if err != nil {
		t.Fatal(err)
	}
	for region != NoRegionAvailable {
		region, err, _ = reader.NextReadableMemoryRegion(region.Address + uintptr(region.Size))
		if err != nil {
			t.Fatal(err)
		}
	}
}
