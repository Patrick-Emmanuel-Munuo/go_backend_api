package main

import (
	"fmt"
	"runtime"
	"os"
	"os/exec"
	"log"
	"time"
	"syscall"
)

func main() {
	// Get the number of CPUs available
	numCPUs := runtime.NumCPU()
	fmt.Printf("Number of CPUs is %d at pid = %d\n", numCPUs, os.Getpid())

	// Start a new goroutine for each CPU
	for i := 0; i < numCPUs; i++ {
		go startWorker(i)
	}

	// Keep the main process alive to manage the workers
	select {}
}

func startWorker(id int) {
	// Simulate a worker's task
	fmt.Printf("Worker %d started, pid = %d\n", id, os.Getpid())

	// Simulating some work by sleeping
	time.Sleep(time.Second * 5)

	// If the worker exits (simulated), restart it
	fmt.Printf("Worker %d (pid = %d) is exiting\n", id, os.Getpid())
	restartWorker(id)
}

func restartWorker(id int) {
	// Simulate restarting a worker (you can use exec.Command to run a new process if needed)
	fmt.Printf("Starting another worker for ID %d\n", id)
	cmd := exec.Command(os.Args[0]) // Run the same Go program again
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Keep the new process in the same process group
	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to restart worker:", err)
	}
	cmd.Wait()
}
