package memsearch

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type MemoryRegion struct {
	address uintptr
	size    uint
}

type Process interface {
	Close() error
	NextReadableMemoryRegion(address uintptr) (MemoryRegion, error)
	ReadMemory(address uintptr, size uint) ([]byte, error)
	CopyMemory(address uintptr, buffer []byte) error
}

type process struct {
	pid          uint
	mapsFilepath string
}

type mapInfo struct {
	start uintptr
	end   uintptr
}

func OpenProcess(pid uint) (Process, error) {
	var result process
	result.pid = pid
	result.mapsFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	// trying to open the maps file, only to see if that gives an error
	f, err := os.Open(result.mapsFilepath)
	if err != nil {
		return nil, err
	}
	f.Close()
	return result, nil
}

func (p process) Close() error {
	return nil
}

// NextReadableMemoryRegion should return a MemoryRegion with address inside or,
// if that's impossible, the next readable MemoryRegion
func (p process) NextReadableMemoryRegion(address uintptr) (MemoryRegion, error) {
	mapsFile, err := os.Open(p.mapsFilepath)
	if err != nil {
		return nil, err
	}
	defer mapsFile.Close()

	name, err := nameByPID(p.pid)
	if err != nil {
		return nil, err
	}

	mappedAddresses, err := getMappedAddresses(mapsFile, name)
	if err != nil {
		return nil, err
	}

	mappedRegion, err := nextReadableMappedRegion(address, mappedAddresses)
	if err != nil {
		return nil, err
	}

	if mappedRegion != nil {
		size := uint(mappedRegion.end - mappedRegion.start)
		return MemoryRegion{mappedRegion.start, size}, nil
	}
	return nil, nil
}

func nextReadableMappedRegion(address uintptr, mappedAddresses []mapInfo) (mapInfo, error) {
	for _, mapinfo := range mappedAddresses {
		if address > mapinfo.start && address < mapinfo.end {
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
	return nil, nil
}

func getMappedAddresses(mapsFile File, name string) ([]mapInfo, error) {
	res := make([]mapInfo, 0)
	scanner := bufio.NewScanner(mapsFile)
	goals := []string{"[heap]", "[stack]", name} // we want to look into the binary memory, its heap and its stack

	for scanner.Scan() {
		line := scanner.Text()

		items := strings.Split(line, " ")
		if len(items) <= 1 {
			continue
		}
		if stringInSlice(items[len(items)-1], goals) {
			fields := strings.Split(items[0], "-")
			start, _ := strconv.ParseInt(fields[0], 16, 64) // TODO: check this. start and end are now uintptr
			end, _ := strconv.ParseInt(fields[1], 16, 64)   // idem
			info := mapInfo{start: start, end: end}
			res = append(res, info)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (p process) ReadMemory(address uintptr, size uint) ([]byte, error) {
	return nil, nil
}

func (p process) CopyMemory(address uintptr, buffer []byte) error {
	var size int
	size = len(buffer)
	buffer = p.ReadMemory(address, size)
	return nil
}

func nameByPID(pid uint) (string, error) {
	// inside /proc/[pid]/stat is the name of the binary between parentheses as second word
	// should be pathByPID (because we need to compare with the full path)
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if a == b {
			return true
		}
	}
	return false
}
