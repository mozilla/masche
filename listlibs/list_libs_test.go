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

	ls, softerrs, err := FindProcWithLib(r)
	fmt.Println(softerrs)
	fmt.Println(ls)
}
