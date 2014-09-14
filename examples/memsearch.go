// This is an example program that shows the usage of the memsearch package.
//
// With this program you can:
//   - Search for a string in the memory of a process with a given PID
//   - Print an arbitrary amount of bytes from the process memory.
package main

import (
	"flag"
	"fmt"
	"github.com/mozilla/masche/memsearch"
	"log"
)

var (
	pid    = flag.Int("pid", 0, "Process id to analyze")
	needle = flag.String("s", "Find This!", "String to search for")
	addr   = flag.Int("addr", 0x0, "Process Address to read")
	size   = flag.Int("n", 4, "Amount of bytes to read")
)

func main() {
	flag.Parse()

	if *addr != 0 {
		if err := memsearch.PrintMemory(uint(*pid), uint64(*addr), *size); err != nil {
			log.Fatal(err)
		}
		return
	}

	if ok, err := memsearch.MemoryGrep(uint(*pid), []byte(*needle)); err != nil {
		log.Fatal(err)
	} else if ok {
		fmt.Println("Found")
	} else {
		fmt.Println("Not Found")
	}
}
