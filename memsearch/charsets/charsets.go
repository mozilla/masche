// This package contains charset-releated function
package charsets

import (
	"encoding/binary"
	"fmt"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

type Charset int

const (
	// Utf-8 needs no alignment variants as it is self-syncrhonizing
	Utf8 Charset = iota

	// Utf-16 starting at even addresses
	Utf16EvenAligned

	// Utf-16 starting at odd addresses
	Utf16OddAligned

	// Utf-32 starting at addresses ≡ 0 (mod 4)
	Utf32Mod0Aligned

	// Utf-32 starting at addresses ≡ 1 (mod 4)
	Utf32Mod1Aligned

	// Utf-32 starting at addresses ≡ 2 (mod 4)
	Utf32Mod2Aligned

	// Utf-32 starting at addresses ≡ 3 (mod 4)
	Utf32Mod3Aligned

	// Windows 1252 needs no aligmen varians because it's a single-byte charset
	Windows1252
)

// An slice with all the supported charsets
var SupportedCharsets = []Charset{Utf8 /* Utf16OddAligned, Utf16EvenAligned*/}

// This rune represents an error while decoding.
// Note that it's used by packages as unicode/utf16 and unicode/utf8 to signal the same.
var runeError = utf8.RuneError

// This type represents a function for decoding the a rune at the begenning if a buffer of bytes. It returns the rune
// and the number of bytes consumed.  It must return runeError if the first size bytes don't form an valid rune.
// Note: utf8.DecodeRune already has this type.
type decodeRuneFunction func(buffer []byte) (r rune, size int)

// Utf-16 decodeRuneFunction
func decodeUtf16Rune(buffer []byte) (r rune, size int) {
	if len(buffer) < 2 {
		return runeError, len(buffer)
	}

	runeBuffer := []uint16{binary.LittleEndian.Uint16(buffer[:2])}
	runes := utf16.Decode(runeBuffer)

	if runes[0] == unicode.ReplacementChar {
		if len(buffer) < 4 {
			return runeError, len(buffer)
		}

		runeBuffer = []uint16{binary.LittleEndian.Uint16(buffer[:2]), binary.LittleEndian.Uint16(buffer[2:4])}
		runes = utf16.Decode(runeBuffer)

		if runes[0] == unicode.ReplacementChar {
			return runeError, 2
		}

		return runes[0], 4
	} else {
		return runes[0], 2
	}
}

// Searches the next string in a buffer starting at startAddress and returns it as a utf-8 string.
// It also the address where the string starts, and the number of bytes consumed from the buffer.
// If no string is found this function returns an empty string after consuming all the buffer.
//
// This function interprets the bytes as runes by calling to decodeFunc, so the charset used dependes on which
// decodeRuneFunction is passed to it.
func getNextString(buffer []byte, startAddress uintptr, decodeFunc decodeRuneFunction) (utf8String string,
	stringStartAddress uintptr, consumedBytes uint) {

	// Skip invalid bytes
	for r, s := decodeFunc(buffer); r == runeError; r, s = decodeFunc(buffer) {
		startAddress += uintptr(s)
		consumedBytes += uint(s)
		buffer = buffer[s:]

		if len(buffer) == 0 {
			return "", startAddress, consumedBytes
		}
	}

	// Add valid characters to string
	for r, s := decodeFunc(buffer); r != runeError; r, s = decodeFunc(buffer) {
		utf8String += string(r)
		consumedBytes += uint(s)

		buffer = buffer[s:]
		if len(buffer) == 0 {
			break
		}
	}

	return utf8String, startAddress, consumedBytes
}

func GetNextString(charset Charset, buffer []byte, startAddress uintptr) (utf8String string,
	stringStartAddress uintptr, consumedBytes uint, err error) {

	switch charset {
	case Utf8:
		utf8String, stringStartAddress, consumedBytes = getNextString(buffer, startAddress, utf8.DecodeRune)
		return

	case Utf16EvenAligned:
		bytesSkipped := startAddress % 2
		utf8String, stringStartAddress, consumedBytes = getNextString(buffer[bytesSkipped:], startAddress+bytesSkipped,
			decodeUtf16Rune)
		return

	case Utf16OddAligned:
		bytesSkipped := (startAddress + 1) % 2
		utf8String, stringStartAddress, consumedBytes = getNextString(buffer[bytesSkipped:], startAddress+bytesSkipped,
			decodeUtf16Rune)

		return
	}

	err = fmt.Errorf("Unrecognized charset")
	return
}
