//TODO: Test CopyMemory and WalkMemory. Specially WalkMemory with different buffer sizes, check that they don't
// overlap, and it doesn't skips memory chunks.
package memaccess

import (
	"github.com/mozilla/masche/test"
	"testing"
)

func TestNewProcessMemoryReader(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	reader, err, softerrors := NewProcessMemoryReader(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
}

func TestManuallyWalk(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	reader, err, softerrors := NewProcessMemoryReader(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	var region MemoryRegion
	region, err, softerrors = reader.NextReadableMemoryRegion(0)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if region == NoRegionAvailable {
		t.Error("No starting region returned")
	}

	previousRegion := region
	for region != NoRegionAvailable {
		region, err, softerrors = reader.NextReadableMemoryRegion(region.Address + uintptr(region.Size))
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}

		if region != NoRegionAvailable && region.Address < previousRegion.Address+uintptr(previousRegion.Size) {
			t.Error("Returned region is not after the previous one.")
		}

		previousRegion = region
	}
}
