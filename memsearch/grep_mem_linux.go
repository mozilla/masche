package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type mapInfo struct {
	start int64
	end   int64
}

type memManagerData struct {
	pos int64
	buf []byte
	memory *os.File
}

//TODO: write memManager
//initializeMemManager returns the memManagerData structure needed to make the subsecuent calls to readAtMemManager
func initializeMemManager(region mapInfo, memory *os.File, length uint) (memManagerData, error){
	res := memManagerData{pos: 0, buf: , memory: memory}
}

//readAtMemManager 
func readAtMemManager(memManagerData, buf []byte, pos uint){
	
}

// mappedAddresses gives the stack and heap addresses for a given pid
func mappedAddresses(pid uint) ([]mapInfo, error) {

	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := make([]mapInfo, 0)
	scanner := bufio.NewScanner(f)
	name, err := nameByPID(pid)
	if err != nil {
		return nil, err
	}
	goals := []string{"[heap]", "[stack]", name} // we want to look into the binary memory, its heap and its stack

	for scanner.Scan() {
		line := scanner.Text()

		items := strings.Split(line, " ")
		if len(items) <= 1 {
			continue
		}
		if stringInSlice(items[len(items)-1], goals) {
			fields := strings.Split(items[0], "-")
			start, _ := strconv.ParseInt(fields[0], 16, 64)
			end, _ := strconv.ParseInt(fields[1], 16, 64)
			info := mapInfo{start: start, end: end}
			res = append(res, info)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
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

func MemoryGrep(pid uint, str []byte) (bool, error) {
	maps, err := mappedAddresses(pid)
	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem")
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	res := false
	for _, region := range maps {
		found, err := searchString(str, region, f)
		if err != nil {
			return false, err
		}
		res = res || found
	}
	return res, nil
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

func searchString(str []byte, region mapInfo, memory *os.File) (bool, error) {
	length = int64(math.Min(512, len(str))) // our buffer will be 512 bytes long unless the string we're looking for is bigger
	buf := make([]byte, length)
	mmanager := initializeMemManager(region, memory, len(str))
	pos = region.start
	for pos < region.end-length {
		_, err := readAtMemManager(mmanager, buf, pos)
		if err != nil {
			return false, err
		}
		if areEqual(str, buf) {
			return true, nil
		}
		pos += length
	}
	return false, nil
}

func areEqual(s1 []byte, s2 []byte) bool {
	for index, _ := range s1 {
		if s1[index] != s2[index] {
			return false
		}
	}
	return true
}

// test main
func main() {
	var pid uint
	var str []byte
	fmt.Scanf("%d", &pid)
	name, _ := nameByPID(pid)
	fmt.Println(name)

	fmt.Scanf("%s", &str)
	res, err := MemoryGrep(pid, str)
	if err != nil {
		fmt.Println("Error", err)
	}
	if res {
		fmt.Println("Encontrado")
	}
}
