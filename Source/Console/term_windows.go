//go:build windows

package Console

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	enableLineInput                 = 0x0002
	enableEchoInput                 = 0x0004
	enableProcessedInput            = 0x0001
	enableVirtualTerminalProcessing = 0x0004
)

func consoleMode(f *os.File) (uint32, bool) {
	var mode uint32
	r, _, _ := procGetConsoleMode.Call(f.Fd(), uintptr(unsafe.Pointer(&mode)))
	return mode, r != 0
}

func setConsoleMode(f *os.File, mode uint32) bool {
	r, _, _ := procSetConsoleMode.Call(f.Fd(), uintptr(mode))
	return r != 0
}

func enterRawMode() (func(), bool) {
	outMode, ok := consoleMode(os.Stdout)
	if !ok {
		return nil, false
	}
	inMode, ok := consoleMode(os.Stdin)
	if !ok {
		return nil, false
	}
	if !setConsoleMode(os.Stdout, outMode|enableVirtualTerminalProcessing) {
		return nil, false
	}
	setConsoleMode(os.Stdin, inMode&^(enableLineInput|enableEchoInput|enableProcessedInput))
	return func() {
		setConsoleMode(os.Stdin, inMode)
		setConsoleMode(os.Stdout, outMode)
	}, true
}

var procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

type consoleInfo struct {
	size, cursor             [2]int16
	attributes               uint16
	left, top, right, bottom int16
	maxSize                  [2]int16
}

func termSize() (int, int) {
	var info consoleInfo
	r, _, _ := procGetConsoleScreenBufferInfo.Call(os.Stdout.Fd(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 80, 30
	}
	return int(info.right-info.left) + 1, int(info.bottom-info.top) + 1
}
