package memaccess

import (
	"bufio"
	"fmt"
	"github.com/mozilla/masche/process"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func mapsFilePathFromPid(pid uint) string {
	return filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
}

func memFilePathFromPid(pid uint) string {
	return filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
}

func nextReadableMemoryRegion(p process.Process, address uintptr) (region MemoryRegion, harderror error,
	softerrors []error) {

	mapsFile, harderror := os.Open(mapsFilePathFromPid(p.Pid()))
	if harderror != nil {
		return
	}
	defer mapsFile.Close()

	region = MemoryRegion{}
	scanner := bufio.NewScanner(mapsFile)

	for scanner.Scan() {
		line := scanner.Text()
		items := splitMapsEntry(line)

		if len(items) != 6 {
			return region, fmt.Errorf("Unrecognised maps line: %s", line), softerrors
		}

		start, end, err := parseMemoryLimits(items[0])
		if err != nil {
			return region, err, softerrors
		}

		if end <= address {
			continue
		}

		// Skip vsyscall as it can't be read. It's a special page mapped by the kernel to accelerate some syscalls.
		if items[5] == "[vsyscall]" {
			continue
		}

		// Check if memory is unreadable
		if items[1][0] == '-' {

			// If we were already reading a region this will just finish it. We only report the softerror when we
			// were actually trying to read it.
			if region.Address != 0 {
				return region, nil, softerrors
			}

			softerrors = append(softerrors, fmt.Errorf("Unreadable memory %s", items[0]))
			continue
		}

		size := uint(end - start)

		// Begenning of a region
		if region.Address == 0 {
			region = MemoryRegion{Address: start, Size: size}
			continue
		}

		// Continuation of a region
		if region.Address+uintptr(region.Size) == start {
			region.Size += size
			continue
		}

		// This map is outside the current region, so we are ready
		return region, nil, softerrors
	}

	// No region left
	if err := scanner.Err(); err != nil {
		return NoRegionAvailable, err, softerrors
	}

	// The last map was a valid region, so it was not closed by an invalid/non-contiguous one and we have to return it
	if region.Address > 0 {
		return region, harderror, softerrors
	}

	return NoRegionAvailable, nil, softerrors
}

func copyMemory(p process.Process, address uintptr, buffer []byte) (harderror error, softerrors []error) {
	mem, harderror := os.Open(memFilePathFromPid(p.Pid()))

	if harderror != nil {
		harderror := fmt.Errorf("Error while reading %d bytes starting at %x: %s", len(buffer), address, harderror)
		return harderror, softerrors
	}
	defer mem.Close()

	bytes_read, harderror := mem.ReadAt(buffer, int64(address))
	if harderror != nil {
		harderror := fmt.Errorf("Error while reading %d bytes starting at %x: %s", len(buffer), address, harderror)
		return harderror, softerrors
	}

	if bytes_read != len(buffer) {
		return fmt.Errorf("Could not read the entire buffer"), softerrors
	}

	return nil, softerrors
}

//Parses the memory limits of a mapping as found in /proc/PID/maps
func parseMemoryLimits(limits string) (start uintptr, end uintptr, err error) {
	fields := strings.Split(limits, "-")
	start64, err := strconv.ParseUint(fields[0], 16, 64)
	if err != nil {
		return 0, 0, err
	}
	start = uintptr(start64)

	end64, err := strconv.ParseUint(fields[1], 16, 64)
	if err != nil {
		return 0, 0, err
	}
	end = uintptr(end64)

	return
}

// splitMapsEntry splits a line of the maps files returning a slice with an element for each of its parts.
func splitMapsEntry(entry string) []string {
	res := make([]string, 0, 6)
	for i := 0; i < 5; i++ {
		if strings.Index(entry, " ") != -1 {
			res = append(res, entry[0:strings.Index(entry, " ")])
			entry = entry[strings.Index(entry, " ")+1:]
		} else {
			res = append(res, entry, "")
			return res
		}
	}
	res = append(res, strings.TrimLeft(entry, " "))
	return res
}
