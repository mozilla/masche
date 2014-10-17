// This is an example program that shows the usage of the memsearch package.
//
// With this program you can:
//   - Search for a string in the memory of a process with a given PID
//   - Print an arbitrary amount of bytes from the process memory.
package main

import (
	"flag"
	"github.com/mozilla/masche/memsearch"
	"log"
)

var (
	pid    = flag.Int("pid", 0, "Process id to analyze")
	prnt   = flag.Bool("print", false, "Print information")
	needle = flag.String("s", "Find This!", "String to search for")
	addr   = flag.Int("addr", 0x0, "Process Address to read")
	size   = flag.Int("n", 4, "Amount of bytes to read")
)

func main() {
	flag.Parse()

	p, err := memsearch.OpenProcess(uint(*pid))
	if err != nil {
		log.Fatal(err)
	}

	if *prnt {
		buf := make([]byte, *size)
		err := p.CopyMemory(uintptr(*addr), buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(buf))
		return
	}

	address, found, err := memsearch.FindNext(p, uintptr(*addr), []byte(*needle))
	if err != nil {
		log.Fatal(err)
	} else if found {
		log.Printf("Found in address: %x\n", address)
	}
}
