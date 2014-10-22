package memsearch

import (
	"../memaccess"
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

func TestOpenProcess(t *testing.T) {
	cmd := exec.Command("../test/tools/test.exe")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	pid := uint(cmd.Process.Pid)
	p, err := memaccess.OpenProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
}

func TestSearchInOtherProcess(t *testing.T) {
	//TODO(mvanotti): Right now the command is hardcoded. We should decide how to fix this.
	cmd := exec.Command("../test/tools/test.exe")

	childout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer childout.Close()

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Wait until the process writes something to stdout, so we know it has initialized all its memory.
	if read, err := childout.Read(make([]byte, 1)); err != nil || read != 1 {
		t.Fatal(err)
	}

	pid := uint(cmd.Process.Pid)
	p, err := memaccess.OpenProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	for i, buf := range buffersToFind {
		_, found, err := FindNext(p, 0, buf)
		if err != nil {
			t.Fatal(err)
		} else if !found {
			t.Fatalf("memoryGrep failed for case %d, the following buffer should be found: %+v", i, buf)
		}
	}

}

func testFindString(t *testing.T) {
	pid := uint(os.Getpid())

	p, err := memaccess.OpenProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	_, found, err := FindNext(p, 0, needle)
	if err != nil {
		t.Fatal(err)
	} else if !found {
		t.Fatalf("memoryGrep failed, searching for %s, should be True", needle)
	}
}
