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
	region, err, softerrors = reader.NextReadableMemoryRegion(0)
	printSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if region == NoRegionAvailable {
		t.Error("No starting region returned")
	}

	previousRegion := region
	for region != NoRegionAvailable {
		region, err, softerrors = reader.NextReadableMemoryRegion(region.Address + uintptr(region.Size))
		printSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}

		if region != NoRegionAvailable && region.Address < previousRegion.Address+uintptr(previousRegion.Size) {
			fmt.Println(previousRegion)
			fmt.Println(region)
			t.Error("Returned region is not after the previous one.")
		}

		previousRegion = region
	}
}
