// This programs executes another program and injects code in it.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Hello, World")

	// Start process
	cmd := exec.Command("./waiter")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Put it to sleep
	p := cmd.Process
	err = p.Signal(syscall.SIGSTOP)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Process Started and Stopped, PID: ", p.Pid)

	// In order to get EIP and write memory we use PTRACE
	// we need to attach ptrace to the process and detach
	// before running it again.
	var regs syscall.PtraceRegs

	if err = syscall.PtraceAttach(p.Pid); err != nil {
		log.Fatal(err)
	}
	if err = syscall.PtraceGetRegs(p.Pid, &regs); err != nil {
		log.Fatal(err)
	}

	WriteCode(p.Pid, regs.Rip)

	if err = syscall.PtraceDetach(p.Pid); err != nil {
		log.Fatal(err)
	}
	log.Println("PTrace Detached")

	err = p.Signal(syscall.SIGCONT)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Process Continued")

	time.Sleep(time.Second * 10)
	p.Kill()
}

// WriteCode writes a small payload into a process at a given address.
func WriteCode(pid int, addr uint64) {
	buf := []byte{0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x48, 0xb8, 0x4d, 0x61, 0x72, 0x63, 0x6f, 0x0a, 0x0a, 0x00, 0x50, 0xbf, 0x01, 0x00, 0x00, 0x00, 0x48, 0x89, 0xe6, 0xba, 0x07, 0x00, 0x00, 0x00, 0xb8, 0x01, 0x00, 0x00, 0x00, 0x0f, 0x05, 0xb8, 0x3c, 0x00, 0x00, 0x00, 0xbf, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x05}

	c, err := syscall.PtracePokeData(pid, uintptr(addr), buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Written %d/%d bytes at offset %x", c, len(buf), addr)
}

// PrintNextBytes prints amount of bytes in hexa starting at the process' rip.
func PrintNextBytes(pid, amount int) {
	var regs syscall.PtraceRegs
	if err := syscall.PtraceGetRegs(pid, &regs); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, amount)
	_, err := syscall.PtracePeekData(pid, uintptr(regs.Rip), buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Next at %x:\n% x", regs.Rip, buf)
}
