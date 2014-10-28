//TODO: Test WalkMemory, specially WalkMemory with different buffer sizes, check that they don't
// overlap, and it doesn't skips memory chunks.
package memaccess

import (
	"fmt"
	"github.com/mozilla/masche/test"
	"os"
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

func TestCopyMemory(t *testing.T) {
	cmd, err := test.LaunchTestCaseAndWaitForInitialization()
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

	min_region_size := uint(os.Getpagesize() + 100) // one page plus something

	for region.Size < min_region_size {
		if region == NoRegionAvailable {
			t.Fatal("We couldn't find a region of %d bytes", min_region_size)
		}

		region, err, softerrors = reader.NextReadableMemoryRegion(region.Address + uintptr(region.Size))
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}

	buffers := [][]byte{
		make([]byte, 2),
		make([]byte, os.Getpagesize()),
		make([]byte, min_region_size),
	}

	for _, buffer := range buffers {
		// Valid read
		err, softerrors = reader.CopyMemory(region.Address, buffer)
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Error(fmt.Sprintf("Couldn't read %d bytes from region", len(buffer)))
		}

		// Crossing boundaries
		err, softerrors = reader.CopyMemory(region.Address+uintptr(region.Size)-uintptr(len(buffer)/2), buffer)
		test.PrintSoftErrors(softerrors)
		if err == nil {
			t.Error(fmt.Sprintf("Read %d bytes inbetween regions", len(buffer)))
		}

		// Entirely outside region
		err, softerrors = reader.CopyMemory(region.Address+uintptr(region.Size), buffer)
		test.PrintSoftErrors(softerrors)
		if err == nil {
			t.Error(fmt.Sprintf("Read %d bytes after the region", len(buffer)))
		}
	}

}

func memoryRegionsOverlap(region1 MemoryRegion, region2 MemoryRegion) bool {
	region1End := region1.Address + uintptr(region1.Size)
	region2End := region2.Address + uintptr(region2.Size)

	if region2.Address >= region1.Address {
		if region2.Address < region1End {
			return true
		}
	} else {
		if region2End <= region1End {
			return true
		}
	}

	return false
}

func TestWalkMemoryDoesntOverlapTheBuffer(t *testing.T) {
	cmd, err := test.LaunchTestCaseAndWaitForInitialization()
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

	pageSize := uint(os.Getpagesize())
	bufferSizes := []uint{1024, pageSize, pageSize + 100, pageSize * 2, pageSize*2 + 123}
	for _, size := range bufferSizes {

		lastRegion := MemoryRegion{}
		err, softerrors = WalkMemory(reader, 0, size, func(address uintptr, buffer []byte) (keepSearching bool) {
			currenRegion := MemoryRegion{Address: address, Size: uint(len(buffer))}
			if memoryRegionsOverlap(lastRegion, currenRegion) {
				t.Errorf("Regions overlap while reading %d at a time: %v %v", size, lastRegion, currenRegion)
				return false
			}

			lastRegion = currenRegion
			return true
		})
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}
}
