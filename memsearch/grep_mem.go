package memsearch

import (
	"bytes"
	"github.com/mozilla/masche/memaccess"
)

// Finds for the first occurrence of needle in the ProcessMemoryReader starting at a given address.
// If the needle is found the first argument will be true and the second one will contain it's address in the other
// process' address space.
//
// TODO(alcuadrado): This doesn't support inter-region reads. Two buffers are needed for this.
func FindNext(reader memaccess.ProcessMemoryReader, address uintptr, needle []byte) (found bool, foundAddress uintptr,
	harderror error, softerrors []error) {

	const min_buffer_size = uint(4096)
	buffer_size := min_buffer_size
	if uint(len(needle)) > buffer_size {
		buffer_size = uint(len(needle))
	}

	foundAddress = uintptr(0)
	found = false
	harderror, softerrors = memaccess.WalkMemory(reader, address, buffer_size,
		func(address uintptr, buf []byte) (keepSearching bool) {
			i := bytes.Index(buf, needle)
			if i == -1 {
				return true
			}

			foundAddress = address + uintptr(i)
			found = true
			return false
		})

	return
}
