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

func openProcessImpl(pid uint) (Process, error) {
	var result process
	result.pid = pid
	result.mapsFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	result.memFilepath = filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
	// trying to open the maps file, only to see if gives an error
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
		return MemoryRegion{}, err
	}
	defer mapsFile.Close()

	path, err := pathByPID(p.pid)
	if err != nil {
		return MemoryRegion{}, err
	}

	mappedAddresses, err := getMappedAddresses(mapsFile, path)
	if err != nil {
		return MemoryRegion{}, err
	}

	mappedRegion, err := nextReadableMappedRegion(address, mappedAddresses)
	if err != nil {
		return MemoryRegion{}, err
	}

	if mappedRegion.start != 0 {
		size := uint(mappedRegion.end - mappedRegion.start)
		return MemoryRegion{mappedRegion.start, size}, nil
	}
	return MemoryRegion{}, nil
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

func (p process) ReadMemory(address uintptr, size uint) ([]byte, error) {
	buffer := make([]byte, size)
	err := p.CopyMemory(address, buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (p process) CopyMemory(address uintptr, buffer []byte) error {
	mem, err := os.Open(p.memFilepath)
	if err != nil {
		//TODO(laski): add address to error string
		return err
	}
	defer mem.Close()
	_, err = mem.ReadAt(buffer, int64(address))
	if err != nil {
		return err
	}
	return nil
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
