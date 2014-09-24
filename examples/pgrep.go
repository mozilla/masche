// This program provides a functionality similar to what `pgrep` does un Linux:
// it lists all the processes whose name matches a given regexp.
package main

import (
	"flag"
	"fmt"
	"github.com/mozilla/masche/listlibs"
	"log"
	"regexp"
)

var reg = flag.String("r", ".*", "Regular Expression to use.")

func main() {
	flag.Parse()

	r, err := regexp.Compile(*reg)
	if err != nil {
		log.Fatal(err)
	}

	ps, _, err := listlibs.GrepProcesses(r)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range ps {
		fmt.Printf("Process: %s\nPid: %d\n\n", p.Filename, p.Pid)
	}
}
