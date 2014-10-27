package memaccess

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type process struct {
	pid          uint
	mapsFilepath string
	memFilepath  string
}

type mapInfo struct {
	start uintptr
	end   uintptr
}

func newProcessMemoryReaderImpl(pid uint) (process, error, []error) {
	var result process
	result.pid = pid
	result.mapsFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	result.memFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
	// trying to open the maps file, only to see if gives an error
	f, harderror := os.Open(result.mapsFilepath)
	softerrors := make([]error, 0)
	if harderror != nil {
		return process{}, harderror, softerrors
	}
	f.Close()
	return result, nil, softerrors
}

func (p process) Close() (error, []error) {
	return nil, make([]error, 0)
}

// NextReadableMemoryRegion should return a MemoryRegion with address inside or,
// if that's impossible, the next readable MemoryRegion
func (p process) NextReadableMemoryRegion(address uintptr) (MemoryRegion, error, []error) {
	mapsFile, harderror := os.Open(p.mapsFilepath)
	softerrors := make([]error, 0)
	if harderror != nil {
		return MemoryRegion{}, harderror, softerrors
	}
	defer mapsFile.Close()

	path, harderror := pathByPID(p.pid)
	if harderror != nil {
		return MemoryRegion{}, harderror, softerrors
	}

	mappedAddresses, harderror := getMappedAddresses(mapsFile, path)
	if harderror != nil {
		return MemoryRegion{}, harderror, softerrors
	}

	mappedRegion, harderror := nextReadableMappedRegion(address, mappedAddresses)
	//TODO: ignore non readable mapped regions and add a softerror
	if harderror != nil {
		return NoRegionAvailable, harderror, softerrors
	}

	if mappedRegion.start != 0 {
		size := uint(mappedRegion.end - mappedRegion.start)
		return MemoryRegion{mappedRegion.start, size}, nil, softerrors
	}
	return NoRegionAvailable, nil, softerrors
}

func nextReadableMappedRegion(address uintptr, mappedAddresses []mapInfo) (mapInfo, error) {
	for _, mapinfo := range mappedAddresses {
		if mapinfo.start <= address && address < mapinfo.end {
			return mapinfo, nil
		}
	}
	// there's no mapped region with address inside it
	// I should return the next one
	for _, mapinfo := range mappedAddresses {
		if address < mapinfo.start {
			return mapinfo, nil
		}
	}
	// there's no mapped region with address inside it and no next region
	return mapInfo{}, nil
}

func getMappedAddresses(mapsFile *os.File, path string) ([]mapInfo, error) {
	res := make([]mapInfo, 0)
	scanner := bufio.NewScanner(mapsFile)
	goals := []string{"[heap]", "[stack]", path} // we want to look into the binary memory, its heap and its stack

	for scanner.Scan() {
		line := scanner.Text()

		items := strings.Split(line, " ")
		if len(items) <= 1 {
			continue
		}
		if stringInSlice(items[len(items)-1], goals) {
			fields := strings.Split(items[0], "-")
			start64, err := strconv.ParseUint(fields[0], 16, 64)
			if err != nil {
				return nil, err
			}
			end64, err := strconv.ParseUint(fields[1], 16, 64)
			if err != nil {
				return nil, err
			}
			start := uintptr(start64)
			end := uintptr(end64)
			info := mapInfo{start: start, end: end}
			res = append(res, info)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (p process) CopyMemory(address uintptr, buffer []byte) (error, []error) {
	mem, harderror := os.Open(p.memFilepath)
	softerrors := make([]error, 0)
	if harderror != nil {
		//TODO(laski): add address to error string
		return harderror, softerrors
	}
	defer mem.Close()
	bytes_read, harderror := mem.ReadAt(buffer, int64(address))
	if bytes_read != len(buffer) {
		return fmt.Errorf("Coul not read the entire buffer"), softerrors
	}
	if harderror != nil {
		return harderror, softerrors
	}
	return nil, softerrors
}

func nameByPID(pid uint) (string, error) {
	// inside /proc/[pid]/stat is the name of the binary between parentheses as second word
	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "stat")
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)

	scanner.Scan() // discard first word
	scanner.Scan()

	res := scanner.Text()
	res = strings.TrimPrefix(res, "(")
	res = strings.TrimSuffix(res, ")")

	return res, nil
}

func pathByPID(pid uint) (string, error) {
	// the file /proc/[pid]/exe is a link to the binary
	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "exe")
	res, err := os.Readlink(path)
	if err != nil {
		return "", err
	}
	return res, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if a == b {
			return true
		}
	}
	return false
}

func areEqual(s1 []byte, s2 []byte) bool {
	for index, _ := range s1 {
		if s1[index] != s2[index] {
			return false
		}
	}
	return true
}
