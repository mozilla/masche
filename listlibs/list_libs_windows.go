package listlibs

// #cgo LDFLAGS: -lpsapi
// #include "list_libs_windows.h"
import "C"

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"unsafe"
)

type ModuleInfo struct {
	filename   string
	baseAddr   uintptr
	size       uint32
	entryPoint uintptr
}

type byPid []uint

func (s byPid) Len() int {
	return len(s)
}
func (s byPid) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byPid) Less(i, j int) bool {
	return s[i] < s[j]
}

func listProcesses() ([]uint, error) {
	r := C.getAllPids()
	defer C.EnumProcessesResponse_Free(r)
	if r.error != 0 {
		return nil, fmt.Errorf("getAllPids failed with error: %d", r.error)
	}

	pids := make([]uint, r.length)

	// We use this to access C arrays without doing manual pointer arithmetic.
	cpids := *(*[]C.DWORD)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(r.pids)),
			Len:  int(r.length),
			Cap:  int(r.length)}))
	for i, _ := range pids {
		pids[i] = uint(cpids[i])
	}

	sort.Sort(byPid(pids))
	return pids, nil
}

func listModules(pid uint) ([]ModuleInfo, error) {
	r := C.getModules(C.DWORD(pid))
	defer C.EnumProcessModulesResponse_Free(r)
	if r.error != 0 {
		return nil, fmt.Errorf("getModules failed with error: %d", r.error)
	}

	mods := make([]ModuleInfo, r.length)

	// We use this to access C arrays without doing manual pointer arithmetic.
	cmods := *(*[]C.ModuleInfo)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(r.modules)),
			Len:  int(r.length),
			Cap:  int(r.length)}))

	for i, _ := range mods {
		mods[i] = ModuleInfo{
			filename:   C.GoString(cmods[i].filename),
			baseAddr:   uintptr(cmods[i].info.lpBaseOfDll),
			size:       uint32(cmods[i].info.SizeOfImage),
			entryPoint: uintptr(cmods[i].info.EntryPoint),
		}
	}

	return mods, nil
}

func HasLibrary(r *regexp.Regexp, pid uint) (bool, error) {
	mods, err := listModules(pid)
	if err != nil {
		return false, err
	}

	for _, m := range mods {
		if r.MatchString(m.filename) {
			return true, nil
		}
	}

	return false, nil
}

func FindProcWithLib(r *regexp.Regexp) ([]uint, error) {
	pids, err := listProcesses()
	if err != nil {
		return nil, err
	}

	var res []uint
	for _, pid := range pids {
		if has, err := HasLibrary(r, pid); err != nil {
			log.Println(err)
			continue
		} else if has {
			res = append(res, pid)
		}
	}

	return res, nil
}
