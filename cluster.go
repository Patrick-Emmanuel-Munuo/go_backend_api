package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

func main() {
	numCPUs := runtime.NumCPU()
	fmt.Printf("Master PID %d: Starting %d workers\n", os.Getpid(), numCPUs)

	for i := 0; i < numCPUs; i++ {
		go startWorker(i)
	}

	select {}
}

func startWorker(id int) {
	for {
		fmt.Printf("Starting worker %d\n", id)
		cmd := exec.Command("./vartrick-server") // Run the built binary directly
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to start worker %d: %v\n", id, err)
			time.Sleep(time.Second)
			continue
		}

		pid := cmd.Process.Pid
		fmt.Printf("Worker %d started with PID %d\n", id, pid)

		err = cmd.Wait()
		if err != nil {
			log.Printf("Worker %d (PID %d) exited with error: %v\n", id, pid, err)
		} else {
			fmt.Printf("Worker %d (PID %d) exited normally\n", id, pid)
		}

		fmt.Printf("Restarting worker %d after 1 second...\n", id)
		time.Sleep(time.Second)
	}
}
