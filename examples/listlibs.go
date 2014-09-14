// This program can be used to check if any process is running a given dynamic library.
// The -r flag specifies a regexp over the filename of the library, for example:
// ./prueba -r="libc" will match all programs that have the libc loaded as a dynamic library.
package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/mozilla/masche/listlibs"
)

var rstr = flag.String("r", "", "library name regexp")

func main() {
	flag.Parse()

	r, err := regexp.Compile(*rstr)
	if err != nil {
		log.Fatal(err)
	}

	procs, softerrs, err := listlibs.FindProcWithLib(r)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range softerrs {
		fmt.Println(e.Error())
	}

	for _, p := range procs {
		fmt.Println(p)
	}

}
