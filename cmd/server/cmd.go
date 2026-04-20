package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func cmdStart() {
	if isRunning() {
		pid, _ := getPid()
		fmt.Printf("Server is already running (PID: %d)\n", pid)
		return
	}

	binPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return
	}

	fmt.Println("Starting server...")

	outFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer outFile.Close()

	cmd := exec.Command(binPath, "run")
	cmd.Stdout = outFile
	cmd.Stderr = outFile
	cmd.SysProcAttr = procAttr()

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

	newPid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(newPid)), 0644); err != nil {
		fmt.Printf("Failed to write PID file: %v\n", err)
		cmd.Process.Kill()
		return
	}

	time.Sleep(1 * time.Second)
	if processExists(newPid) {
		fmt.Printf("Server started (PID: %d)\n", newPid)
	} else {
		fmt.Println("Server failed to start")
		removePidFile()
		printLogTail()
	}
}

func cmdStop() {
	if !isRunning() {
		fmt.Println("Server is not running")
		removePidFile()
		return
	}

	pid, _ := getPid()
	fmt.Printf("Stopping server (PID: %d)...\n", pid)

	process, _ := os.FindProcess(pid)
	_ = signalStop(process)

	waitForProcess(pid, 30*time.Second)

	removePidFile()
	fmt.Println("Server stopped")
}

func cmdRestart() {
	if isRunning() {
		cmdStop()
		time.Sleep(2 * time.Second)
	}
	cmdStart()
}

func cmdReload() {
	if !isRunning() {
		fmt.Println("Server is not running")
		return
	}

	pid, _ := getPid()
	fmt.Printf("Sending SIGUSR1 to reload (PID: %d)...\n", pid)

	process, _ := os.FindProcess(pid)
	if err := signalReload(process); err != nil {
		fmt.Printf("Reload is not supported on this platform: %v\n", err)
		return
	}

	fmt.Println("Reload signal sent")
}

func cmdStatus() {
	if !isRunning() {
		fmt.Println("Server is not running")
		return
	}

	pid, _ := getPid()
	fmt.Printf("Server is running (PID: %d)\n", pid)

	if data, err := os.ReadFile(logFile); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			fmt.Println("\nLast 5 lines of log:")
			start := 0
			if len(lines) > 5 {
				start = len(lines) - 5
			}
			for i := start; i < len(lines); i++ {
				if lines[i] != "" {
					fmt.Println(lines[i])
				}
			}
		}
	}
}

func cmdHotRestart() {
	if !isRunning() {
		fmt.Println("Server is not running, starting fresh...")
		cmdStart()
		return
	}

	oldPid, _ := getPid()

	binPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return
	}

	fmt.Println("Hot restart: starting new process...")

	outFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer outFile.Close()

	cmd := exec.Command(binPath, "run")
	cmd.Stdout = outFile
	cmd.Stderr = outFile
	cmd.SysProcAttr = procAttr()

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

	newPid := cmd.Process.Pid
	time.Sleep(2 * time.Second)

	if processExists(newPid) {
		fmt.Printf("New server started (PID: %d), stopping old (PID: %d)...\n", newPid, oldPid)

		oldProcess, _ := os.FindProcess(oldPid)
		_ = signalStop(oldProcess)

		waitForProcess(oldPid, 30*time.Second)

		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(newPid)), 0644); err != nil {
			fmt.Printf("Failed to update PID file: %v\n", err)
		}

		fmt.Println("Hot restart complete")
	} else {
		fmt.Println("New server failed to start, keeping old one")
		removePidFile()
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(oldPid)), 0644); err != nil {
			fmt.Printf("Failed to restore PID file: %v\n", err)
		}
		printLogTail()
	}
}

func printUsage() {
	fmt.Println(`xin - XinFramework Server Management

Usage: xin <command>

Commands:
  start        Start the server (daemon mode)
  stop         Graceful stop (SIGTERM, 30s timeout)
  restart      Stop then start
  reload       Hot reload (SIGUSR1, no dropped connections)
  hot-restart  Zero-downtime restart (start new, stop old)
  status       Show server status

Examples:
  xin start
  xin status
  xin stop`)
}

func printLogTail() {
	if data, err := os.ReadFile(logFile); err == nil {
		fmt.Println("\nServer log:")
		fmt.Print(string(data))
	}
}
