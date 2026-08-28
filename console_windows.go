//go:build windows

package main

import (
	"os"
	"syscall"
)

var procAttachConsole = syscall.NewLazyDLL("kernel32.dll").NewProc("AttachConsole")

func attachConsole() {
	procAttachConsole.Call(^uintptr(0)) // ATTACH_PARENT_PROCESS = -1

	if h, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE); err == nil && h != 0 {
		os.Stdout = os.NewFile(uintptr(h), "stdout")
	}
	if h, err := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE); err == nil && h != 0 {
		os.Stderr = os.NewFile(uintptr(h), "stderr")
	}
	if h, err := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE); err == nil && h != 0 {
		os.Stdin = os.NewFile(uintptr(h), "stdin")
	}
}
