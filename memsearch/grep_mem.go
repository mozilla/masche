package memsearch

import (
	"bytes"
	"github.com/mozilla/masche/memaccess"
)

// FindNext finds for the first occurrence of needle in the memory of Process ph after the given address.
func FindNext(ph memaccess.Process, address uintptr, needle []byte) (uintptr, bool, error) {
	addr := uintptr(0)
	found := false
	hard, _ := memaccess.WalkMemory(ph, address, 4096,
		func(address uintptr, buf []byte) (keepSearching bool) {
			i := bytes.Index(buf, needle)
			if i == -1 {
				return true
			}

			addr = address + uintptr(i)
			found = true
			return false
		})
	return addr, found, hard
}
