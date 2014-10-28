// This packages contains an interface for accessing other processes' memory.
package memaccess

import "fmt"

// Creates a new ProcessMemoryReader for a given process.
func NewProcessMemoryReader(pid uint) (reader ProcessMemoryReader, harderror error, softerrors []error) {
	return newProcessMemoryReaderImpl(pid)
}

// This interface is used to access a process memory. For getting an instance
// of a type implementing it you must call NewProcessMemoryReader function.
type ProcessMemoryReader interface {
	// Frees the resources attached to this interface.
	Close() (harderror error, softerrors []error)

	// Returns a memory region containing address, or the next readable region
	// after address in case addresss is not in a readable region.
	//
	// If there aren't more regions available the special value
	// NoRegionAvailable is returned.
	NextReadableMemoryRegion(address uintptr) (region MemoryRegion, harderror error, softerrors []error)

	// Fills the entire buffer with memory from the process starting in address (in the process address space).
	// If there is not enough memory to read it returns a hard error. Note that this is not the only hard error it may
	// return though.
	CopyMemory(address uintptr, buffer []byte) (harderror error, softerrors []error)
}

// This struct represents a region of readable contiguos memory of a process.
//
// No readable memory can be available right next to this region, it's maximal
// in its upper bound.
//
// Note that this region is not necessary equivalent to the OS's region, if any.
type MemoryRegion struct {
	Address uintptr
	Size    uint
}

func (m MemoryRegion) String() string {
	return fmt.Sprintf("MemoryRegion[%x-%x)", m.Address, m.Address+uintptr(m.Size))
}

// A centinel value indicating that there is no more regions available.
var NoRegionAvailable MemoryRegion

// This type represents a function used for walking through the memory, see
// WalkMemory for more details.
type WalkFunc func(address uintptr, buf []byte) (keepSearching bool)

// WalkMemory reads all the memory of a process starting at a given address
// reading upto bufSize bytes into a buffer, and calling walkFn with the buffer
// and the start address of the memory in the buffer. If walkFn returns false
// WalkMemory stop reading the memory.
func WalkMemory(reader ProcessMemoryReader, startAddress uintptr, bufSize uint, walkFn WalkFunc) (harderror error,
	softerrors []error) {

	var region MemoryRegion
	region, harderror, softerrors = reader.NextReadableMemoryRegion(startAddress)
	if harderror != nil {
		return
	}

	const max_retries int = 5

	buf := make([]byte, bufSize)
	retries := max_retries

	for region != NoRegionAvailable {

		keepWalking, addr, err, serrs := walkRegion(reader, region, buf, walkFn)
		softerrors = append(softerrors, serrs...)

		if err != nil && retries > 0 {
			// An error occurred: retry using the nearest region to the address that failed.
			retries--
			region, harderror, serrs = reader.NextReadableMemoryRegion(addr)
			softerrors = append(softerrors, serrs...)
			if harderror != nil {
				return
			}

			// if some chunk of this new region was already read we don't want to read it again.
			if region.Address < addr {
				region.Address = addr
			}

			continue
		} else if err != nil {
			// we have exceeded our retries, mark the error as soft error and keep going.
			softerrors = append(softerrors, fmt.Errorf("Retries exceeded on reading %d bytes starting at %x: %s",
				len(buf), addr, err.Error()))
		} else if !keepWalking {
			return
		}

		region, harderror, serrs = reader.NextReadableMemoryRegion(region.Address + uintptr(region.Size))
		softerrors = append(softerrors, serrs...)
		if harderror != nil {
			return
		}
		retries = max_retries
	}
	return
}

// This function walks through a single memory region calling walkFunc with a given buffer. It always fills as much of
// the buffer as possible before calling walkFunc, but it never calls it with overlaped memory sections.
//
// If the buffer cannot be filled a hard error is returned with the starting address of the chunk of memory that could
// not be read. If no harderror is returned errorAddress must be ignored.
//
// If any of the calls to walkFn returns false, this function inmediatly returns, with keepWalking set to false and no
// hard error.
func walkRegion(reader ProcessMemoryReader, region MemoryRegion, buf []byte, walkFn WalkFunc) (keepWalking bool,
	errorAddress uintptr, harderror error, softerrors []error) {

	softerrors = make([]error, 0)
	keepWalking = true

	remaningBytes := uintptr(region.Size)
	for addr := region.Address; remaningBytes > 0; addr += uintptr(len(buf)) {
		err, serrs := reader.CopyMemory(addr, buf)
		softerrors = append(softerrors, serrs...)

		if err != nil {
			harderror = err
			errorAddress = addr
			return
		}

		keepWalking = walkFn(addr, buf)
		if !keepWalking {
			return
		}

		remaningBytes -= uintptr(len(buf))
		if remaningBytes < uintptr(len(buf)) {
			buf = buf[:remaningBytes]
		}
	}

	return
}
