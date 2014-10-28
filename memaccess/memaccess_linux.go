package memaccess

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type process struct {
	pid          uint
	mapsFilepath string
	memFilepath  string
}

func newProcessMemoryReaderImpl(pid uint) (process, error, []error) {
	var result process
	result.pid = pid
	result.mapsFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	result.memFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
	softerrors := make([]error, 0)

	// trying to open the maps file, only to see if we have enought privileges
	f, harderror := os.Open(result.mapsFilepath)
	if harderror != nil {
		return process{}, harderror, softerrors
	}
	f.Close()

	return result, nil, softerrors
}

func (p process) Close() (error, []error) {
	return nil, make([]error, 0)
}

func (p process) NextReadableMemoryRegion(address uintptr) (region MemoryRegion, harderror error, softerrors []error) {
	// fmt.Printf("\n\n\nNextReadableMemoryRegion %x\n", address)
	softerrors = make([]error, 0)

	mapsFile, harderror := os.Open(p.mapsFilepath)
	if harderror != nil {
		return
	}
	defer mapsFile.Close()

	region = MemoryRegion{}
	scanner := bufio.NewScanner(mapsFile)
	splitBySpacesRegexp := regexp.MustCompile("\\s+")

	for scanner.Scan() {
		line := scanner.Text()
		items := splitBySpacesRegexp.Split(line, -1)

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
			if region.Address == 0 {
				return region, nil, softerrors
			}

			softerrors = append(softerrors, fmt.Errorf("Unreadable memory %s - address %x", items[0], address))
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
		return region, err, softerrors
	}

	return NoRegionAvailable, nil, softerrors
}

func (p process) CopyMemory(address uintptr, buffer []byte) (error, []error) {
	mem, harderror := os.Open(p.memFilepath)
	softerrors := make([]error, 0)
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
		return fmt.Errorf("Coul not read the entire buffer"), softerrors
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
