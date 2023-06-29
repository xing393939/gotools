// 参考：https://bye04t46if.execute-api.ap-southeast-1.amazonaws.com/default/Web?url=https://medium.com/golangspec/making-debugger-for-golang-part-i-53124284b7c8
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

func setPC(pid int, pc uint64) {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	regs.SetPC(pc)
	err = syscall.PtraceSetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
}

func cont(pid int) {
	err := syscall.PtraceCont(pid, 0)
	if err != nil {
		log.Fatal("cont ", err, pid)
	}
}

func setBreakpoint(pid int, breakpoint uint64) byte {
	original := make([]byte, 1)
	_, err := syscall.PtracePeekData(pid, uintptr(breakpoint), original)
	if err != nil {
		log.Fatal("setBreakpoint ", err)
	}
	_, err = syscall.PtracePokeData(pid, uintptr(breakpoint), []byte{0xCC})
	if err != nil {
		log.Fatal("setBreakpoint ", err)
	}
	return original[0]
}

func clearBreakpoint(pid int, breakpoint uint64, ori byte) {
	original := make([]byte, 1)
	original[0] = ori
	_, err := syscall.PtracePokeData(pid, uintptr(breakpoint), original)
	if err != nil {
		log.Fatal("clearBreakpoint ", err)
	}
}

func getCurrentRegs(pid int) syscall.PtraceRegs {
	var regs syscall.PtraceRegs
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatal(err)
	}
	return regs
}

func main() {
	if len(os.Args) != 2 {
		println("usage: gocallstack /path/to/exe")
		return
	}
	cmd := exec.Command(os.Args[1])
	cmd.Args = os.Args[1:]
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	pid := cmd.Process.Pid
	log.Printf("pid:%d %v\n", pid, err)

	// 打开go二进制文件
	file, err := elf.Open(os.Args[1])
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
	breakpointMap := make(map[uint64]byte)
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
		if strings.Contains(symbol.Name, "_cgo_") {
			continue
		}
		if !strings.Contains(symbol.Name, "_rt0_") {
			continue
		}
		original := setBreakpoint(pid, symbol.Value)
		breakpointMap[symbol.Value] = original
	}

	var ws syscall.WaitStatus
	cont(pid)
	waitPid, waitErr := syscall.Wait4(pid, &ws, syscall.WALL, nil)
	for waitErr == nil {
		if ws == 0x057f {
			regs := getCurrentRegs(waitPid)
			name := getSymbol(symbols, regs.Rip)
			log.Printf("pid:%d %x %s %x\n", waitPid, ws, name, regs.Rip)

			breakpoint := regs.Rip - 1
			clearBreakpoint(waitPid, breakpoint, breakpointMap[breakpoint])
			setPC(waitPid, breakpoint)
		}

		cont(waitPid)
		waitPid, waitErr = syscall.Wait4(pid, &ws, syscall.WALL, nil)
	}
	log.Printf("%s %x \n", waitErr.Error(), ws)
}
