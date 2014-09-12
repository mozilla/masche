package memsearch

// #include "procMems_windows.h"
// #cgo CFLAGS: -std=c99
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

func PrintMemory(pid uint, addr uint64, size int) error {
	minfo := C.GetMemoryInformation(C.DWORD(pid))
	defer C.MemoryInformation_Free(minfo)
	if minfo.error != 0 {
		return fmt.Errorf("GetMemoryInformation failed with error %d\n", minfo.error)
	}

	C.PrintMemory(minfo.hndl, C.PVOID(uintptr(addr)), C.SIZE_T(size))
	return nil
}

func MemoryGrep(pid uint, buf []byte) (bool, error) {
	minfo := C.GetMemoryInformation(C.DWORD(pid))
	defer C.MemoryInformation_Free(minfo)
	if minfo.error != 0 {
		return false, fmt.Errorf("GetMemoryInformation failed with error %d", minfo.error)
	}

	cinfo := *(*[]C.MEMORY_BASIC_INFORMATION)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(minfo.info)),
			Len:  int(minfo.length),
			Cap:  int(minfo.length)}))

	cbuf := C.CString(string(buf))
	clen := C.int(len(buf))

	results := make(chan bool)

	for _, v := range cinfo {
		go func(v C.MEMORY_BASIC_INFORMATION) {
			res := C.FindInRange(minfo.hndl, v, cbuf,clen)

			results <- int(res) != 0
		}(v)
	}

	done := 0
	for {
	select {
		case v := <- results:
			done += 1
		if v {
			return true, nil
		}
		if done == int(minfo.length) {
			return false, nil
		}
	}
	}
	return false, nil
}
