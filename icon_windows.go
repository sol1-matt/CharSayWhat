//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	shell32            = syscall.NewLazyDLL("shell32.dll")
	procSendMessageW   = user32.NewProc("SendMessageW")
	procExtractIconExW = shell32.NewProc("ExtractIconExW")
)

const (
	wmSetIcon = 0x0080
	iconSmall = 0
	iconBig   = 1
)

func setWindowIconFromExe(hwnd uintptr) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exePtr, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return
	}
	var big, small uintptr
	procExtractIconExW.Call(
		uintptr(unsafe.Pointer(exePtr)),
		0,
		uintptr(unsafe.Pointer(&big)),
		uintptr(unsafe.Pointer(&small)),
		1,
	)
	if small != 0 {
		procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, small)
	}
	if big != 0 {
		procSendMessageW.Call(hwnd, wmSetIcon, iconBig, big)
	}
}
