package memsearch

import (
	"bytes"
	"github.com/mozilla/masche/memaccess"
)

// FindNext finds for the first occurrence of needle in the memory of ProcessMemoryReader reader after the given address.
func FindNext(reader memaccess.ProcessMemoryReader, address uintptr, needle []byte) (uintptr, bool, error) {
	addr := uintptr(0)
	found := false
	hard, _ := memaccess.WalkMemory(reader, address, 4096,
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
