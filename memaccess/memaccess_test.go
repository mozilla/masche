package memaccess

import (
	"github.com/mozilla/masche/process"
	"github.com/mozilla/masche/test"
	"os"
	"testing"
)

func TestManuallyWalk(t *testing.T) {
	cmd, err := test.LaunchTestCase()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	proc, softerrors, err := process.OpenFromPid(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	var region MemoryRegion
	region, softerrors, err = NextReadableMemoryRegion(proc, 0)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if region == NoRegionAvailable {
		t.Error("No starting region returned")
	}

	previousRegion := region
	for region != NoRegionAvailable {
		region, softerrors, err = NextReadableMemoryRegion(proc, region.Address+uintptr(region.Size))
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
	proc, softerrors, err := process.OpenFromPid(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	var region MemoryRegion
	region, softerrors, err = NextReadableMemoryRegion(proc, 0)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if region == NoRegionAvailable {
		t.Error("No starting region returned")
	}

	minRegionSize := uint(os.Getpagesize() + 100) // one page plus something

	for region.Size < minRegionSize {
		if region == NoRegionAvailable {
			t.Fatal("We couldn't find a region of %d bytes", minRegionSize)
		}

		region, softerrors, err = NextReadableMemoryRegion(proc, region.Address+uintptr(region.Size))
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}

	buffers := [][]byte{
		make([]byte, 2),
		make([]byte, os.Getpagesize()),
		make([]byte, minRegionSize),
	}

	for _, buffer := range buffers {
		// Valid read
		softerrors, err = CopyMemory(proc, region.Address, buffer)
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Errorf("Couldn't read %d bytes from region", len(buffer))
		}

		// Crossing boundaries
		softerrors, err = CopyMemory(proc, region.Address+uintptr(region.Size)-uintptr(len(buffer)/2), buffer)
		test.PrintSoftErrors(softerrors)
		if err == nil {
			t.Errorf("Read %d bytes inbetween regions", len(buffer))
		}

		// Entirely outside region
		softerrors, err = CopyMemory(proc, region.Address+uintptr(region.Size), buffer)
		test.PrintSoftErrors(softerrors)
		if err == nil {
			t.Errorf("Read %d bytes after the region", len(buffer))
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
	proc, softerrors, err := process.OpenFromPid(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	pageSize := uint(os.Getpagesize())
	bufferSizes := []uint{1024, pageSize, pageSize + 100, pageSize * 2, pageSize*2 + 123}
	for _, size := range bufferSizes {

		lastRegion := MemoryRegion{}
		softerrors, err = WalkMemory(proc, 0, size, func(address uintptr, buffer []byte) (keepSearching bool) {
			currentRegion := MemoryRegion{Address: address, Size: uint(len(buffer))}
			if memoryRegionsOverlap(lastRegion, currentRegion) {
				t.Errorf("Regions overlap while reading %d at a time: %v %v", size, lastRegion, currentRegion)
				return false
			}

			lastRegion = currentRegion
			return true
		})
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestWalkRegionReadsEntireRegion(t *testing.T) {
	cmd, err := test.LaunchTestCaseAndWaitForInitialization()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	proc, softerrors, err := process.OpenFromPid(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	pageSize := uint(os.Getpagesize())
	bufferSizes := []uint{1024, pageSize, pageSize + 100, pageSize * 2, pageSize*2 + 123}

	var region MemoryRegion
	region, softerrors, err = NextReadableMemoryRegion(proc, 0)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	if region == NoRegionAvailable {
		t.Error("No starting region returned")
	}

	minRegionSize := bufferSizes[len(bufferSizes)-1]
	for region.Size < minRegionSize {
		if region == NoRegionAvailable {
			t.Fatal("We couldn't find a region of %d bytes", minRegionSize)
		}

		region, softerrors, err = NextReadableMemoryRegion(proc, region.Address+uintptr(region.Size))
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, size := range bufferSizes {
		buf := make([]byte, size)
		readRegion := MemoryRegion{}

		_, _, softerrors, err := walkRegion(proc, region, buf,
			func(address uintptr, buffer []byte) (keepSearching bool) {
				if readRegion.Address == 0 {
					readRegion.Address = address
					readRegion.Size = uint(len(buffer))
					return true
				}

				readRegionLimit := readRegion.Address + uintptr(readRegion.Size)
				if readRegionLimit != address {
					t.Errorf("walkRegion skept %d bytes starting at %x", address-readRegionLimit,
						readRegionLimit)
					return false
				}

				readRegion.Size += uint(len(buffer))
				return true
			})
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}

		if region != readRegion {
			t.Errorf("%v not entirely read", region)
		}
	}
}

func TestSlidingWalkMemory(t *testing.T) {
	cmd, err := test.LaunchTestCaseAndWaitForInitialization()
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	proc, softerrors, err := process.OpenFromPid(pid)
	test.PrintSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}

	pageSize := uint(os.Getpagesize())
	bufferSizes := []uint{1024, pageSize, pageSize + 100, pageSize * 2, pageSize*2 + 124}
	for _, size := range bufferSizes {
		lastRegion := MemoryRegion{}
		softerrors, err = SlidingWalkMemory(proc, 0, size, func(address uintptr, buffer []byte) (keepSearching bool) {
			currentRegion := MemoryRegion{Address: address, Size: uint(len(buffer))}

			if lastRegion.Address == 0 {
				lastRegion = currentRegion
				return true
			}

			lastRegionLimit := lastRegion.Address + uintptr(lastRegion.Size)
			overlappedBytes := uintptr(0)
			regionIsContigous := false

			if lastRegionLimit > currentRegion.Address {
				overlappedBytes = lastRegionLimit - currentRegion.Address
			} else if lastRegionLimit == currentRegion.Address {
				regionIsContigous = true
			}

			if regionIsContigous {
				if regionIsContigous {
					t.Errorf("Contigous buffer while we are expecting overlapped ones."+
						"buffer size %d - lastRegion %v - currentRegion %v", size, lastRegion, currentRegion)
				}
				return false
			}

			// If the last buffer wasn't read complete the current one can't be overlapped
			if lastRegion.Size != size && overlappedBytes > 0 {
				t.Errorf("Overlapped buffer after non-complete one. "+
					"buffer size %d - lastRegion %v - currentRegion %v", size, lastRegion, currentRegion)
				return false
			}

			// Overlapped bytes should be half of the buffer, or the buffer must came from another region
			if overlappedBytes != uintptr(size/2) && overlappedBytes != 0 {
				t.Errorf("Overlapping buffer by %d bytes. "+
					"buffer size %d - lastRegion %v - currentRegion %v", overlappedBytes, size, lastRegion,
					currentRegion)
				return false
			}

			lastRegion = currentRegion
			return true
		})
		test.PrintSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		}
	}
}
