//go:build windows

package framework

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess        = kernel32.NewProc("OpenProcess")
	procCloseHandle        = kernel32.NewProc("CloseHandle")
	procGetExitCodeProcess = kernel32.NewProc("GetExitCodeProcess")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	STILL_ACTIVE                      = 259
)

func notifySignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,
	}
}

func isRunning() bool {
	pid, err := getPid()
	if err != nil {
		return false
	}
	return processExists(pid)
}

func processExists(pid int) bool {
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		0,
		uintptr(pid),
	)
	if handle == 0 {
		return false
	}
	defer procCloseHandle.Call(handle)

	var exitCode uint32
	procGetExitCodeProcess.Call(handle, uintptr(unsafe.Pointer(&exitCode)))
	return exitCode == STILL_ACTIVE
}
