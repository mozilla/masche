package main 


import (
	"fmt"
	"flag"
	"log"
	"github.com/mozilla/masche/memsearch"
)

var (
	pid = flag.Int("pid", 0, "Process id to analyze")
	needle = flag.String("s", "Find This!", "String to search for")
)

func main() {
	flag.Parse()

	if ok, err := memsearch.MemoryGrep(uint(*pid), []byte(*needle)); err != nil {
		log.Fatal(err)
	} else if ok {
		fmt.Println("Found")
	} else {
		fmt.Println("Not Found")
	}
}