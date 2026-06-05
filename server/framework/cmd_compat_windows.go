//go:build windows

package framework

import (
	"errors"
	"os"
	"syscall"
)

func procAttr() *syscall.SysProcAttr {
	return nil
}

func signalStop(process *os.Process) error {
	return process.Kill()
}

func signalReload(process *os.Process) error {
	return errors.New("SIGUSR1 unavailable")
}
