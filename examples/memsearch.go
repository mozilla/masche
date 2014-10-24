// This is an example program that shows the usage of the memsearch package.
//
// With this program you can:
//   - Search for a string in the memory of a process with a given PID
//   - Print an arbitrary amount of bytes from the process memory.
package main

import (
	"flag"
	"github.com/mozilla/masche/memaccess"
	"github.com/mozilla/masche/memsearch"
	"log"
	"regexp"
)

var (
	pid      = flag.Int("pid", 0, "Process id to analyze")
	prnt     = flag.Bool("print", false, "Print information")
	needle   = flag.String("s", "Find This!", "String to search for")
	addr     = flag.Int("addr", 0x0, "Process Address to read")
	size     = flag.Int("n", 4, "Amount of bytes to read")
	isRegexp = flag.Bool("r", false, "Assume needle is a regexp")
)

func main() {
	flag.Parse()

	reader, err, _ := memaccess.NewProcessMemoryReader(uint(*pid))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	if *prnt {
		buf := make([]byte, *size)
		err, _ := reader.CopyMemory(uintptr(*addr), buf)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(buf))
		return
	}

	if *isRegexp {
		r, err := regexp.Compile(*needle)
		if err != nil {
			log.Fatal(err)
		}

		found, address, err, _ := memsearch.FindNextMatch(reader,
			uintptr(*addr), r)
		if err != nil {
			log.Fatal(err)
		} else if found {
			log.Printf("Found in address: %x\n", address)
		}
		return
	}

	found, address, err, _ := memsearch.FindNext(reader,
		uintptr(*addr), []byte(*needle))
	if err != nil {
		log.Fatal(err)
	} else if found {
		log.Printf("Found in address: %x\n", address)
	}
}
