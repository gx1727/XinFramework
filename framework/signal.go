package framework

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gx1727.com/xin/framework/internal/core/boot"

	"gx1727.com/xin/framework/internal/core/server"
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

	for range ticker.C {
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

func waitForSignal(srv *server.XinServer) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, notifySignals()...)

	sig := <-sigCh
	log.Printf("Received signal: %v", sig)

	if err := srv.Shutdown(30 * time.Second); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	boot.Shutdown()

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
