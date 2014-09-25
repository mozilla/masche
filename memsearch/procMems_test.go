package memsearch

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var needle []byte = []byte("Find This!")

var buffersToFind = [][]byte{
	[]byte{0xc, 0xa, 0xf, 0xe},
	[]byte{0xd, 0xe, 0xa, 0xd, 0xb, 0xe, 0xe, 0xf},
	[]byte{0xb, 0xe, 0xb, 0xe, 0xf, 0xe, 0x0},
}

func TestSearchInOtherProcess(t *testing.T) {
	//TODO(mvanotti): Right now the command is hardcoded. We should decide how to fix this.
	cmd := exec.Command("../test/tools/test.exe")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	fmt.Println("PID: ", pid)
	fmt.Println("My PID: ", os.Getpid())

	for _, buf := range buffersToFind {
		res, err := MemoryGrep(pid, buf)
		if err != nil {
			t.Fatal(err)
		} else if !res {
			t.Fatal("memoryGrep failed, the following buffer should be found", buf)
		}
	}

}

func testFindString(t *testing.T) {
	pid := uint(os.Getpid())

	res, err := MemoryGrep(pid, needle)
	if err != nil {
		t.Fatal(err)
	} else if !res {
		t.Fatalf("memoryGrep failed, searching for %s, should be True", needle)
	}
}
