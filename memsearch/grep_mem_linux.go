package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type mapInfo struct {
	start int64
	end   int64
}

// mappedAddresses gives the stack and heap addresses for a given pid
func mappedAdresses(pid uint) ([]mapInfo, error) {

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
	maps, err := mappedAdresses(pid)
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
	buf := make([]byte, len(str))
	for pos := region.start; pos < region.end-int64(len(str)); pos++ {
		_, err := memory.ReadAt(buf, pos)
		if err != nil {
			return false, err
		}
		if areEqual(str, buf) {
			return true, nil
		}
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
