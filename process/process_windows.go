package process

// #cgo CFLAGS: -std=c99
// #cgo CFLAGS: -DPSAPI_VERSION=1
// #cgo LDFLAGS: -lpsapi
// #include "memaccess.h"
// #include "process_windows.h"
import "C"

type proc uintptr

func getAllPids() (pids []uint, harderror error, softerrors []error) {
	r := C.getAllPids()
	defer C.EnumProcessesResponse_Free(r)
	if r.error != 0 {
		return nil, fmt.Errorf("getAllPids failed with error %d", r.error)
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

	return pids, nil, nil
}

func openFromPid(pid uint) (p Process, harderror error, softerrors []error) {
	//TODO(mvanotti): Complete this with memaccess implementation
	return nil, nil, nil
}

type proc struct {
	hndl C.process_handle_t
	pid  uint
}

func (p proc) Pid() uint {
	return p.pid
}

func (p proc) Name() (name string, harderror error, softerrors []error) {
	//TODO(mvanotti): Complete this with listlibs implementation
}

func openProcessByPid(pid uint) (p Process, harderror error, softerrors []error) {
	var result proc

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

func (p proc) Close() (harderror error, softerrors []error) {
	resp := C.close_process_handle(p.hndl)
	defer C.response_free(resp)
	return getResponseErrors(resp)
}
