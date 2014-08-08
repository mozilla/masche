package main

import (
	"flag"
	"fmt"
	"log"
	"memgator"
	"regexp"
)

var rstr = flag.String("r", "", "library name regexp")

func main() {
	flag.Parse()

	r, err := regexp.Compile(*rstr)
	if err != nil {
		log.Fatal(err)
	}

	procs, err := memgator.FindProcWithLib(r)
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range procs {
		fmt.Println(p)
	}

}
