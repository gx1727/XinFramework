package main

import (
	"fmt"
	"gx1727.com/xin/internal/core/boot"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func getPid() (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func isRunning() bool {
	pid, err := getPid()
	if err != nil {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func removePidFile() {
	os.Remove(pidFile)
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func waitForProcess(pid int, timeout time.Duration) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	waited := 0 * time.Second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !processExists(pid) {
				return
			}
			waited += time.Second
			if waited >= timeout {
				fmt.Println("Force killing...")
				process.Kill()
				time.Sleep(time.Second)
				return
			}
			fmt.Printf("Waiting... (%v/%v)\n", waited, timeout)
		}
	}
}

func waitForSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	sig := <-sigCh
	log.Printf("Received signal: %v", sig)

	boot.GetServer().Shutdown(30 * time.Second)

	log.Printf("Server exited gracefully")
	os.Exit(0)
}

func sdNotifyReady() error {
	if os.Getenv("NOTIFY_SOCKET") == "" {
		return nil
	}

	socketPath := os.Getenv("NOTIFY_SOCKET")
	conn, err := net.Dial("unixgram", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("READY=1"))
	return err
}
