//go:build !windows

package main

import (
	"os"
	"syscall"
)

func procAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func signalStop(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}

func signalReload(process *os.Process) error {
	return process.Signal(syscall.SIGUSR1)
}
