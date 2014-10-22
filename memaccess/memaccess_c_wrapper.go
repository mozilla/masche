// +build windows darwin

package memaccess

// #include "memaccess.h"
// #cgo CFLAGS: -std=c99
import "C"

import (
	"fmt"
	"unsafe"
)

type process struct {
	hndl C.process_handle_t
	pid  uint
}

func openProcessImpl(pid uint) (Process, error) {
	var result process

	resp := C.open_process_handle(C.pid_tt(pid), &result.hndl)
	defer C.response_free(resp)
	if resp.fatal_error != nil {
		return result, fmt.Errorf("open_process_handle failed with error %d: %s",
			resp.fatal_error.error_number,
			C.GoString(resp.fatal_error.description))
	}

	result.pid = pid
	return result, nil

}

func (p process) Close() error {
	resp := C.close_process_handle(p.hndl)
	defer C.response_free(resp)
	if resp.fatal_error != nil {
		return fmt.Errorf("close_process_handle failed with error %d: %s",
			resp.fatal_error.error_number,
			C.GoString(resp.fatal_error.description))
	}

	return nil
}

func (p process) NextReadableMemoryRegion(address uintptr) (MemoryRegion, error) {
	var isAvailable C.bool
	var region C.memory_region_t

	response := C.get_next_readable_memory_region(
		p.hndl,
		C.memory_address_t(address),
		&isAvailable,
		&region)
	defer C.response_free(response)
	return MemoryRegion{uintptr(region.start_address), uint(region.length)}, nil
}

func (p process) ReadMemory(address uintptr, size uint) ([]byte, error) {
	return nil, nil
}

func (p process) CopyMemory(address uintptr, buffer []byte) error {
	buf := unsafe.Pointer(&buffer[0])

	n := len(buffer)
	var bytesRead C.size_t
	resp := C.copy_process_memory(p.hndl,
		C.memory_address_t(address),
		C.size_t(n),
		buf,
		&bytesRead,
	)

	defer C.response_free(resp)

	if resp.fatal_error != nil {
		return fmt.Errorf("copy_process_memory failed with error %d: %s",
			resp.fatal_error.error_number,
			C.GoString(resp.fatal_error.description))
	}
	return nil
}
