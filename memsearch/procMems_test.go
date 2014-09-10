package memsearch

import (
	"os"
	"testing"
)

var needle []byte = []byte("Find This!")

func TestFindString(t *testing.T) {
	pid := uint(os.Getpid())

	res, err := memoryGrep(pid, needle)
	if err != nil {
		t.Fatal(err)
	} else if !res {
		t.Fatalf("memoryGrep failed, searching for %s, should be True", needle)
	}
}
