// This packages contains an interface for accessing other processes' memory.
//TODO(alcuadrado): Add documentation about error handling
package memaccess

//TODO(alcuadrado): Add documentation.
func OpenProcess(pid uint) (Process, error) {
	return openProcessImpl(pid)
}

// This interface is used to access a process memory. For getting an instance
// of a type implementing it you must call OpenProcess function defined below.
type Process interface {
	// Frees the resources attached to this interface.
	Close() error

	// Returns a memory region containing address, or the next readable region
	// after address in case addresss is not in a readable region.
	//
	// If there aren't more regions available the special value
	// NoRegionAvailable is returned.
	NextReadableMemoryRegion(address uintptr) (MemoryRegion, error)

	//TODO(alcuadrado): Add a detailed doc about how this works, specially in
	// corner cases.
	CopyMemory(address uintptr, buffer []byte) error
}

// This struct represents a region of readable contiguos memory of a process.
//
// No readable memory can be available right next to this region, it's maximal
// in its upper bound.
//
// Note that this region is not necessary equivalent to the OS's region, if any.
type MemoryRegion struct {
	address uintptr
	size    uint
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
//
// TODO(alcuadrado): Add documentation about error handling and retries.
func WalkMemory(p Process, startAddress uintptr, bufSize uint, walkFn WalkFunc) (harderror error, softerrors []error) {
	region, harderror := p.NextReadableMemoryRegion(startAddress)
	if harderror != nil {
		return
	}

	const max_retries int = 5

	buf := make([]byte, bufSize)
	retries := max_retries
	softerrors = make([]error, 0)

	for region != NoRegionAvailable {
		keepWalking, addr, err := walkRegion(p, region, buf, walkFn)
		if err != nil && retries > 0 {
			// An error occurred: retry using the nearest region to the address that failed.
			retries--
			region, harderror = p.NextReadableMemoryRegion(addr)
			if harderror != nil {
				return
			}

			continue
		} else if err != nil {
			// we have exceeded our retries, mark the error as soft error and keep going.
			softerrors = append(softerrors, err)
		} else if !keepWalking {
			return
		}

		region, harderror = p.NextReadableMemoryRegion(region.address + uintptr(region.size))
		if harderror != nil {
			return
		}
		retries = max_retries
	}
	return
}

//TODO(mvanotti): change the multiple return value to a specific error with an address field.
//TODO(mvanotti): Add documentation.
//TODO(alcuadrado): Clean this code.
func walkRegion(p Process, region MemoryRegion, buf []byte, walkFn WalkFunc) (bool, uintptr, error) {
	bufSz := uint(len(buf))
	address := region.address

	// We divide the memory region in blocks of bufSz/2 bytes, and for each block (except the last one), we read bufSz bytes.
	// We leave the last block out, so we can later read the last bufSz bytes starting from te end (reading the last block and the remaining)
	// For example, if we have a region with size 17 and bufSize of size 8, we divide it in regions of 4 bytes:
	// 0-3, 4-7, 8-11, 12-15 and the reamining byte is left out: 16
	// we read 0 to 7 (8 bytes), 4 to 11 (8 bytes), and 8 to 15 (8 bytes).
	// after that, we read the last 8 bytes: 9 to 16.
	// in this case, i will take values 0, 1 and 2
	for i := uint(0); i < region.size/(bufSz/2)-1; i, address = i+1, address+uintptr(bufSz/2) {
		err := p.CopyMemory(address, buf)
		if err != nil {
			return false, address, err
		}
		if !walkFn(address, buf) {
			return false, 0, nil
		}
	}

	// Get the remaining part:
	// We have at most bufSize/2 bytes to walk, so we copy the last bufSize bytes to get the window between the last two chunks.
	if region.size%(bufSz/2) != 0 {
		address = region.address + uintptr(region.size-bufSz)
		err := p.CopyMemory(address, buf)
		if err != nil {
			return false, address, err
		}

		return walkFn(address, buf), 0, nil
	}

	return true, 0, nil
}
