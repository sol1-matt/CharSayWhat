//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
	procFreeConsole           = kernel32.NewProc("FreeConsole")
)

const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

// prepareGUI gets the console out of the way before the GUI opens. The exe is
// a console-subsystem binary so shells wait for CLI output (a windowsgui exe
// lets PowerShell print its prompt over the output). It reports whether this
// process should go on to run the GUI.
func prepareGUI() bool {
	var pids [2]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	switch n {
	case 0: // no console at all, e.g. the detached relaunch below
		return true
	case 1: // Windows created a console just for us (Explorer, shortcut): drop it
		procFreeConsole.Call()
		return true
	}

	// The console is shared with a shell. Relaunch detached and exit so the
	// prompt comes straight back instead of blocking until the GUI closes.
	exe, err := os.Executable()
	if err != nil {
		return true
	}
	cmd := exec.Command(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGroup}
	if err := cmd.Start(); err != nil {
		return true // run in place rather than not at all
	}
	return false
}
