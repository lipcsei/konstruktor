package main

import (
	/*
		int isEven(int number) {
		    return number % 2 == 0;
		}
	*/
	"C"
	"fmt"
	"github.com/lipcsei/konstruktor/worker"
	"log"
	"math/big"
	"runtime"
	"strconv"
	"sync"
)

// numTasks defines the total number of tasks to be generated and processed.
const numTasks = 100

func main() {

	// Create channels for tasks and results with a capacity of numTasks.
	tasks := make(chan worker.Task, numTasks)     // The tasks channel is used to send tasks to the workers.
	results := make(chan worker.Result, numTasks) // The results channel is for receiving processed tasks from the workers.

	// quit is a channel used to signal workers to stop processing and exit gracefully.
	// This is particularly useful for terminating workers once all tasks have been processed.
	quit := make(chan struct{})

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
		w := worker.New(workerID, tasks, results, &wg, quit)
		// Start the worker in a new goroutine.
		go w.Start()
	}

	go func() {
		wg.Wait()      // Wait for all workers to finish.
		close(results) // Close the results channel to signal completion of result processing.
		close(quit)    // Close the quit channel as a final step, signaling any remaining workers to terminate.
	}()

	// Collect and print the results. SortResults organizes results into their original order based on task ID.
	for _, result := range worker.SortResults(results, numTasks) {
		if result.Factorial.Cmp(big.NewInt(0)) != 0 {
			runes := []rune(result.Factorial.String())
			lastRune := fmt.Sprintf("%c", runes[len(runes)-1])
			lastDigit, err := strconv.Atoi(lastRune)
			if err != nil {
				log.Println(err)
				continue
			}
			if lastDigit == 0 || (lastDigit != 0 && C.isEven(C.int(lastDigit)) == 1) {
				log.Printf("%d worker finishe the %d. task: %d! = %d The result is an even number. \n", result.WorkerID, result.Task.ID, result.Task.Value, result.Factorial)
			}
		} else {
			log.Printf("%d. task: %d! != %d The computation failed due to a timeout. \n", result.Task.ID, result.Task.Value, result.Factorial)
		}
	}

}
