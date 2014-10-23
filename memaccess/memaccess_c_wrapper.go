// +build windows darwin

package memaccess

// #include "memaccess.h"
// #cgo CFLAGS: -std=c99
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Go representation of an error returned by memaccess.h functions
type osError struct {
	number      int
	description string
}

func (err osError) Error() string {
	return fmt.Sprintf("System error number %d: %s", err.number, err.description)
}

// Tranforms a C.error_t into a osError
func cErrorToOsError(cError C.error_t) osError {
	return osError{
		number:      int(cError.error_number),
		description: C.GoString(cError.description),
	}
}

// Returns the Go representation of the errors present in a C.reponse_t
func getResponseErrors(response *C.response_t) (harderror error, softerrors []error) {
	if response.fatal_error != nil && int(response.fatal_error.error_number) != 0 {
		harderror = cErrorToOsError(*response.fatal_error)
	} else {
		harderror = nil
	}

	softerrorsCount := int(response.soft_errors_count)
	softerrors = make([]error, 0, softerrorsCount)

	cSoftErrorsHeader := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(response.soft_errors)),
		Len:  softerrorsCount,
		Cap:  softerrorsCount,
	}
	cSoftErrors := *(*[]C.error_t)(unsafe.Pointer(&cSoftErrorsHeader))

	for _, cErr := range cSoftErrors {
		softerrors = append(softerrors, cErrorToOsError(cErr))
	}

	return
}

// ProcessMemoryReader implementation
type process struct {
	hndl C.process_handle_t
	pid  uint
}

func newProcessMemoryReaderImpl(pid uint) (reader ProcessMemoryReader, harderror error, softerrors []error) {
	var result process

	resp := C.open_process_handle(C.pid_tt(pid), &result.hndl)
	harderror, softerrors = getResponseErrors(resp)
	C.response_free(resp)

	if harderror == nil {
		result.pid = pid
	} else {
		resp = C.close_process_handle(result.hndl)
		C.response_free(resp)
	}

	return result, harderror, softerrors
}

func (p process) Close() (harderror error, softerrors []error) {
	resp := C.close_process_handle(p.hndl)
	defer C.response_free(resp)
	return getResponseErrors(resp)
}

func (p process) NextReadableMemoryRegion(address uintptr) (region MemoryRegion, harderror error, softerrors []error) {
	var isAvailable C.bool
	var cRegion C.memory_region_t

	response := C.get_next_readable_memory_region(
		p.hndl,
		C.memory_address_t(address),
		&isAvailable,
		&cRegion)
	harderror, softerrors = getResponseErrors(response)
	C.response_free(response)

	if harderror != nil || isAvailable == false {
		return NoRegionAvailable, harderror, softerrors
	}

	return MemoryRegion{uintptr(cRegion.start_address), uint(cRegion.length)}, harderror, softerrors
}

func (p process) CopyMemory(address uintptr, buffer []byte) (harderror error, softerrors []error) {
	buf := unsafe.Pointer(&buffer[0])

	n := len(buffer)
	var bytesRead C.size_t
	resp := C.copy_process_memory(p.hndl,
		C.memory_address_t(address),
		C.size_t(n),
		buf,
		&bytesRead,
	)

	harderror, softerrors = getResponseErrors(resp)
	C.response_free(resp)

	if harderror != nil {
		return
	}

	if len(buffer) != int(bytesRead) {
		harderror = fmt.Errorf("Coul not read the entire buffer")
	}

	return
}
