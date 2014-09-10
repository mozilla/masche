package memsearch

// #include "procMems_windows.h"
// #cgo CFLAGS: -std=c99
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

func memoryGrep(pid uint, buf []byte) (bool, error) {
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

	for _, v := range cinfo {
		res := C.FindInRange(minfo.hndl, v, C.CString(string(buf)), C.int(len(buf)))
		if int(res) != 0 {
			return true, nil
		}
	}

	return false, nil
}
