package memgator

// #cgo LDFLAGS: -lpsapi
// #include "list_libs_windows.h"
import "C"

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"unsafe"
)

type byPid []uint32

func (s byPid) Len() int {
	return len(s)
}
func (s byPid) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byPid) Less(i, j int) bool {
	return s[i] < s[j]
}

func listProcesses() []uint32 {
	r := C.getAllPids()
	if r.error != 0 {
		log.Fatal(r.error)
	}

	defer C.EnumProcessesResponse_Free(r)

	fmt.Println(r.length)
	pids := make([]uint32, r.length)

	for i, _ := range pids {
		//TODO(mvanotti): See if there's a cleaner way to convert the C array to Go.
		address := uintptr(unsafe.Pointer(r.pids)) + unsafe.Sizeof(C.DWORD(0))*uintptr(i)
		pids[i] = *(*uint32)(unsafe.Pointer(address))
	}

	sort.Sort(byPid(pids))
	fmt.Println(pids)
	return pids
}

func listModules(pid uint32) []string {
	r := C.getModules(C.DWORD(pid))
	if r.error != 0 {
		log.Fatal(err)
	}
	defer C.EnumProcessModulesReponse_Free(r)

	fmt.Println(r.length)
	modules := make([]string, r.length)
	for i, _ := range modules {
		address := uintptr(unsafe.Pointer(r.modules)) + unsafe.Sizeof(*C.char)*uintptr(i)
		modules[i] = C.GoString(*(**C.char)(unsafe.Pointer(address)))
	}

	sort.Strings(modules)
	fmt.Println(modules)
	return modules
}

func HasLibrary(r *regexp.Regexp) (bool, error) {
	return true, nil
}

func FindProcWithLib(r *regexp.Regexp) ([]int, error) {
	return nil, nil
}
