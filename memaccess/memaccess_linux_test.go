package memaccess

import (
	"testing"
)

func TestSplitMapsEntry(t *testing.T) {
	var entries = []string{
		"7fb8faf65000-7fb8faf66000 rw-p 00023000 08:01 922969                     /lib/x86_64-linux-gnu/ld-2.19.so",
		"7fb8faf65000-7fb8faf66000 rw-p 00023000 08:01 922969                     /lib/x86_64-linux-gnu/with spaces.so",
		"7fb8faf66000-7fb8faf67000 rw-p 00000000 00:00 0",
		"7fff231a6000-7fff231c7000 rw-p 00000000 00:00 0          [stack]",
		"7fff231a6000-7fff231c7000 rw-p 00000000 00:00 0                          [stack]",
	}

	var results = [][]string{
		[]string{"7fb8faf65000-7fb8faf66000", "rw-p", "00023000", "08:01", "922969",
			"/lib/x86_64-linux-gnu/ld-2.19.so"},
		[]string{"7fb8faf65000-7fb8faf66000", "rw-p", "00023000", "08:01", "922969",
			"/lib/x86_64-linux-gnu/with spaces.so"},
		[]string{"7fb8faf66000-7fb8faf67000", "rw-p", "00000000", "00:00", "0", ""},
		[]string{"7fff231a6000-7fff231c7000", "rw-p", "00000000", "00:00", "0", "[stack]"},
		[]string{"7fff231a6000-7fff231c7000", "rw-p", "00000000", "00:00", "0", "[stack]"},
	}

	for i, entry := range entries {
		splitted := splitMapsEntry(entry)
		if !compareStringSlices(results[i], splitted) {
			t.Error("Error splitting map entry", entry, " - Expected:", results[i], " - Got: ", splitted)
		}
	}
}

func compareStringSlices(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
