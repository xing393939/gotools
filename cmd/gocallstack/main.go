package main

import (
	"debug/elf"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func getBk(symbols []elf.Symbol, funName string) uint64 {
	for _, symbol := range symbols {
		if symbol.Name == funName {
			return symbol.Value
		}
	}
	return 0
}

func getSymbol(symbols []elf.Symbol, ptr uint64) string {
	for _, symbol := range symbols {
		if symbol.Value <= ptr && ptr-symbol.Value <= symbol.Size {
			return symbol.Name
		}
	}
	return "unknow"
}

func setPC(pid int, pc uintptr) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	regs.SetPC(uint64(pc))
	err = syscall.PtraceSetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
}

func cont(pid int) {
	err := syscall.PtraceCont(pid, 0)
	if err != nil {
		log.Fatal("cont ", err)
	}
}

func setBreakpoint(pid int, breakpoint uintptr) byte {
	original := make([]byte, 1)
	_, err := syscall.PtracePeekData(pid, breakpoint, original)
	if err != nil {
		log.Fatal("setBreakpoint ", err)
	}
	_, err = syscall.PtracePokeData(pid, breakpoint, []byte{0xCC})
	if err != nil {
		log.Fatal("setBreakpoint ", err)
	}
	return original[0]
}

func clearBreakpoint(pid int, breakpoint uintptr, ori byte) {
	original := make([]byte, 1)
	original[0] = ori
	_, err := syscall.PtracePokeData(pid, breakpoint, original)
	if err != nil {
		log.Fatal("clearBreakpoint ", err)
	}
}

func getCurrentBk(pid int) uintptr {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	return uintptr(regs.Rip - 1)
}

func main() {
	input := "../hello/hello"
	cmd := exec.Command(input)
	cmd.Args = []string{input}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	log.Printf("State: %v\n", err)

	// 打开go二进制文件
	file, err := elf.Open(input)
	if err != nil {
		log.Println("Error opening file")
	}
	defer file.Close()

	// 获取符号表
	symbols, err := file.Symbols()
	if err != nil {
		log.Println("Error getting symbols")
	}

	// 设置所有的断点
	pid := cmd.Process.Pid
	bkMap := make(map[uintptr]byte)
	for _, symbol := range symbols {
		if symbol.Size == 0 {
			continue
		}
		if strings.Contains(symbol.Name, "runtime.") {
			continue
		}
		if strings.Contains(symbol.Name, "internal/") {
			continue
		}
		if strings.Contains(symbol.Name, "$f64.") {
			continue
		}
		if strings.Contains(symbol.Name, "_cgo_") {
			continue
		}
		if strings.Contains(symbol.Name, "_rt0_") {
			continue
		}

		breakpoint := uintptr(symbol.Value)
		original := setBreakpoint(pid, breakpoint)
		bkMap[breakpoint] = original
	}

	var ws syscall.WaitStatus
	cont(pid)
	_, err = syscall.Wait4(pid, &ws, syscall.WALL, nil)
	for err == nil {
		breakpoint := getCurrentBk(pid)
		name := getSymbol(symbols, uint64(breakpoint))
		log.Printf("%s %x\n", name, breakpoint)

		clearBreakpoint(pid, breakpoint, bkMap[breakpoint])
		setPC(pid, breakpoint)

		cont(pid)
		_, err = syscall.Wait4(pid, &ws, syscall.WALL, nil)
	}
	log.Println(err.Error())
}
