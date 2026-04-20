//go:build !windows

package main

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
