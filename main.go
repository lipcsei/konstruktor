package main

import (
	"fmt"
	"github.com/lipcsei/konstruktor/worker"
	"runtime"
	"sync"
)

const numTasks = 100

func main() {

	tasks := make(chan worker.Task, numTasks)
	results := make(chan worker.Result, numTasks)

	go worker.GenerateTasks(numTasks, tasks)

	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU() + 1
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		w := worker.New(i, tasks, results, &wg)
		go w.Start()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for _, result := range worker.SortResults(results, numTasks) {
		fmt.Printf("%d. task: %d! = %d\n", result.Task.ID, result.Task.Value, result.Factorial)
	}

}
