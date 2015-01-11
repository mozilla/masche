package memsearch

import (
	"bytes"
	"github.com/mozilla/masche/memaccess"
	"github.com/mozilla/masche/process"
	"regexp"
)

// Finds for the first occurrence of needle in the Process starting at a given address.
// If the needle is found the first argument will be true and the second one will contain it's address in the other
// process' address space.
func FindNext(p process.Process, address uintptr, needle []byte) (found bool, foundAddress uintptr,
	harderror error, softerrors []error) {

	const min_buffer_size = uint(4096)
	buffer_size := min_buffer_size
	if uint(len(needle)) > buffer_size {
		buffer_size = uint(len(needle))
	}

	foundAddress = uintptr(0)
	found = false
	harderror, softerrors = memaccess.SlidingWalkMemory(p, address, buffer_size,
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

func FindNextMatch(p process.Process, address uintptr, r *regexp.Regexp) (found bool, foundAddress uintptr,
	harderror error, softerrors []error) {

	const buffer_size = uint(4096)

	foundAddress = uintptr(0)
	found = false
	harderror, softerrors = memaccess.SlidingWalkMemory(p, address, buffer_size,
		func(address uintptr, buf []byte) (keepSearching bool) {
			loc := r.FindIndex(buf)
			if loc == nil {
				return true
			}

			foundAddress = address + uintptr(loc[0])
			found = true
			return false
		})

	return
}
