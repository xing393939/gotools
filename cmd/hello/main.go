package main

import (
	"net"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

type epollevent struct {
	events uint32
	data   [8]byte // unaligned uintptr
}

//go:noinline
func EpollWait(epfd int, events []epollevent, msec int) (n int, err error) {
	var r0 uintptr
	var _p0 = unsafe.Pointer(&events[0])
	r0, _, err = syscall.Syscall6(syscall.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(_p0), uintptr(len(events)), uintptr(msec), 0, 0)
	return int(r0), err
}

func EpollCtl(epfd int, op int, fd int, event *epollevent) (err error) {
	_, _, err = syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if err == syscall.Errno(0) {
		err = nil
	}
	return err
}

func Wait(ep int) (err error) {
	// init
	var msec, n = -1, 0
	events := make([]epollevent, 1)
	// wait
	for {
		n, err = EpollWait(ep, events, msec)
		if err != nil && err != syscall.EINTR {
			return err
		}
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		msec = 0
	}
}

func parseFD(ln net.Listener) int {
	var file *os.File
	switch netln := ln.(type) {
	case *net.TCPListener:
		file, _ = netln.File()
	case *net.UnixListener:
		file, _ = netln.File()
	default:
		return 0
	}

	return int(file.Fd())
}

func main() {
	listener, _ := net.Listen("tcp", "localhost:6380")
	ep, _ := syscall.EpollCreate1(0)

	var evt epollevent
	var op int
	op, evt.events = syscall.EPOLL_CTL_ADD, syscall.EPOLLIN|syscall.EPOLLRDHUP|syscall.EPOLLERR
	_ = EpollCtl(ep, op, parseFD(listener), &evt)
	_ = Wait(ep)
}
