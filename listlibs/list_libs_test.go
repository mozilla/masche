package listlibs

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFindProcWithLib(t *testing.T) {
	r, err := regexp.Compile("libc")
	if err != nil {
		t.Fatal(err)
	}

	ls, err := FindProcWithLib(r)
	fmt.Println(ls)
}
