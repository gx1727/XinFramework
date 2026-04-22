//go:build !windows

package framework

import (
	"os"
	"syscall"
)

func notifySignals() []os.Signal {
	return []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
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
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
