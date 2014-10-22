package memaccess

//TODO(alcuadrado): Add documentation.
type MemoryRegion struct {
	address uintptr
	size    uint
}

//TODO(alcuadrado): Add documentation.
type Process interface {
	Close() error
	NextReadableMemoryRegion(address uintptr) (MemoryRegion, error)
	ReadMemory(address uintptr, size uint) ([]byte, error)
	CopyMemory(address uintptr, buffer []byte) error
}

//TODO(alcuadrado): Add documentation.
type WalkFunc func(address uintptr, buf []byte) (keepSearching bool)

//TODO(alcuadrado): Add documentation.
var emptyRegion MemoryRegion

//TODO(alcuadrado): Add documentation.
func WalkMemory(p Process, startAddress uintptr, bufSize uint, walkFn WalkFunc) (harderror error, softerrors []error) {
	region, harderror := p.NextReadableMemoryRegion(startAddress)
	if harderror != nil {
		return
	}

	const max_retries int = 5

	buf := make([]byte, bufSize)
	retries := max_retries
	softerrors = make([]error, 0)

	for region != emptyRegion {
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
