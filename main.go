package main

import (
	"fmt"
	"github.com/lipcsei/konstruktor/worker"
	"runtime"
	"sync"
)

// numTasks defines the total number of tasks to be generated and processed.
const numTasks = 100

func main() {

	// Create channels for tasks and results with a capacity of numTasks.
	tasks := make(chan worker.Task, numTasks)
	results := make(chan worker.Result, numTasks)

	// Start a goroutine to generate tasks
	go worker.GenerateTasks(numTasks, tasks)

	// wg is a WaitGroup to wait for all worker goroutines to finish processing.
	var wg sync.WaitGroup

	// numWorkers is a determined the number of workers based on the number of CPU cores + 1.
	numWorkers := runtime.NumCPU() + 1
	for workerID := 0; workerID < numWorkers; workerID++ {
		// Increment the WaitGroup counter for each worker.
		wg.Add(1)
		// Initialize a new worker.
		w := worker.New(workerID, tasks, results, &wg)
		// Start the worker in a new goroutine.
		go w.Start()
	}

	// Start a goroutine to close the results channel once all workers have finished.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and print the results. SortResults organizes results into their original order.
	for _, result := range worker.SortResults(results, numTasks) {
		fmt.Printf("%d. task: %d! = %d\n", result.Task.ID, result.Task.Value, result.Factorial)
	}
}
