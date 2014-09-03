package main

import (
    "bufio"
    "strconv"
    "os"
    "strings"
    "fmt"
    "path/filepath"
)

type mapInfo struct {
    start int64
    end int64
}

func mappedAdresses(pid uint) ([]mapInfo, error) {
    path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "maps") 
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    res := make([]mapInfo, 0)

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()

        items := strings.Split(line, " ")
        if len(items) <= 1 {
            continue
        }
        if (items[len(items)-1] == "[heap]" || items[len(items)-1] == "[stack]"){
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

func MemoryGrep(pid uint, str []byte) (bool, error) {
    maps, err := mappedAdresses(pid)
    path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "mem") 
    f, err := os.Open(path)
    if err != nil {
        return false, err
    }
    defer f.Close()

    buf := make([]byte, len(str))
    for _, info := range maps{
        for pos := info.start; pos < info.end; pos++ {
            f.ReadAt(buf , pos)
            if areEqual(str, buf){
                return true, nil
            }
        }
    }
    return false, nil
}

func areEqual(s1 []byte, s2 []byte) bool {
    for index, _ := range s1{
        if s1[index] != s2[index]{
            return false
        }
    }
    return true
}

func main(){
    pid := uint(12981)
    str := []byte{0}
    maps, _ := mappedAdresses(pid)
    for _, info := range maps{
        fmt.Printf("%x - %x \n", info.start, info.end)
    }
    res, _ := MemoryGrep(pid, str)
    if res {
        fmt.Println("encontrado")
    }
}