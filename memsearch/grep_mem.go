package memsearch

import (
	"bytes"
	"fmt"
	"github.com/mozilla/masche/memaccess"
	"github.com/mozilla/masche/memsearch/charsets"
	"regexp"
)

// Finds for the first occurrence of needle in the ProcessMemoryReader starting at a given address.
// If the needle is found the first argument will be true and the second one will contain it's address in the other
// process' address space.
func FindNext(reader memaccess.ProcessMemoryReader, address uintptr, needle []byte) (found bool, foundAddress uintptr,
	harderror error, softerrors []error) {

	const min_buffer_size = uint(4096)
	buffer_size := min_buffer_size
	if uint(len(needle)) > buffer_size {
		buffer_size = uint(len(needle))
	}
	if buffer_size%2 != 0 {
		buffer_size += 1
	}

	foundAddress = uintptr(0)
	found = false
	harderror, softerrors = memaccess.SlidingWalkMemory(reader, address, buffer_size,
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

func FindNextMatch(reader memaccess.ProcessMemoryReader, address uintptr, r *regexp.Regexp,
	possibleCharsets []charsets.Charset) (found bool, foundAddress uintptr, harderror error, softerrors []error) {

	const buffer_size = uint(4096) * 4

	if possibleCharsets == nil {
		possibleCharsets = charsets.SupportedCharsets
	}

	foundAddress = uintptr(0)
	found = false
	harderror, softerrors = memaccess.SlidingWalkMemory(reader, address, buffer_size,
		func(address uintptr, buf []byte) (keepSearching bool) {
			previousAddress := uintptr(0)
			for _, charset := range possibleCharsets {
				currentBuffer := buf
				currentAddress := address

				for len(currentBuffer) != 0 {
					if currentAddress <= previousAddress {
						fmt.Printf("WTF! %x <= %x\n", currentAddress, previousAddress)
					}
					previousAddress = currentAddress
					str, startAddress, consumedBytes, err := charsets.GetNextString(charset, currentBuffer,
						currentAddress)

					if err != nil {
						softerrors = append(softerrors, fmt.Errorf("Error decoding string: %s", err))
						break
					}

					loc := r.FindStringIndex(str)
					if loc != nil {
						foundAddress = startAddress + uintptr(loc[0])
						found = true
						return false
					}

					currentBuffer = currentBuffer[consumedBytes:]
					currentAddress += uintptr(consumedBytes)
				}
			}

			return true
		})

	return
}
