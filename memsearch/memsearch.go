package memsearch

import (
	"bytes"
	"github.com/mozilla/masche/memaccess"
	"github.com/mozilla/masche/process"
	"regexp"
)

// FindBytesSequence finds for the first occurrence of needle in the Process starting at a given address (in the
// process address space). If the needle is found the first argument will be true and the second one will contain it's
// address.
func FindBytesSequence(p process.Process, address uintptr, needle []byte) (found bool, foundAddress uintptr,
	softerrors []error, harderror error) {

	const minBufferSize = uint(4096)
	bufferSize := minBufferSize
	if uint(len(needle)) > bufferSize {
		bufferSize = uint(len(needle))
	}

	foundAddress = uintptr(0)
	found = false
	softerrors, harderror = memaccess.SlidingWalkMemory(p, address, bufferSize,
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

// FindRegexpMatch finds the first match of r in the process memory. This function works as FindFindBytesSequence
// but instead of searching for a literal bytes sequence it uses a regexp. It tries to match the regexp in the memory
// as is, not interpreting it as any charset in particular.
func FindRegexpMatch(p process.Process, address uintptr, r *regexp.Regexp) (found bool, foundAddress uintptr,
	softerrors []error, harderror error) {

	const bufferSize = uint(4096)

	foundAddress = uintptr(0)
	found = false
	softerrors, harderror = memaccess.SlidingWalkMemory(p, address, bufferSize,
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
