package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Global variables for storing processing times and for synchronization
var (
	processingTimes           []time.Duration // Slice to store the processing times of tasks
	processingTimesLock       sync.Mutex      //  Mutex to synchronize access to the processingTimes slice
	maxProcessingTimesToTrack = 20            //  Maximum number of processing times to keep track of for average calculation
)

var simulateDelay func()

// Task represents a unit of work to process.
type Task struct {
	// ID is an unique identifier that also determines result sorting order
	ID int
	// Value specifies the number for which the factorial is to be calculated.
	Value int64
}

// Result represents the outcome of a Task, including its factorial result
type Result struct {
	Task      Task
	Factorial *big.Int
}

// GenerateTasks generates a specified number of tasks with random values and sends them on a channel
func GenerateTasks(numTasks int, tasks chan<- Task) {
	for i := 0; i < numTasks; i++ {
		// Generate a random number between 3 and 1000
		randomNumber := int64(rand.Intn(998) + 3)

		task := Task{
			ID:    i,
			Value: randomNumber,
		}
		tasks <- task
	}
	close(tasks) // Signal to processors that there are no more tasks
}

// Worker processes tasks and sends results on a channel.
func Worker(id int, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		startTime := time.Now()

		if simulateDelay != nil {
			simulateDelay()
		}

		result := Factorial(task.Value)

		processingTime := time.Since(startTime)

		averageTime := calculateAverageProcessingTime()

		processingTimesLock.Lock()
		if len(processingTimes) >= maxProcessingTimesToTrack {
			processingTimes = processingTimes[1:] // Az első elem eltávolítása
		}
		processingTimes = append(processingTimes, processingTime)
		processingTimesLock.Unlock()

		// Calculate the allowed time threshold as 10% above the average time
		allowedTimeThreshold := averageTime + (averageTime / 10)

		// Check if the processing time exceeds the allowed time threshold
		if processingTime > 0 && averageTime > 0 && processingTime > allowedTimeThreshold {
			result = big.NewInt(0) // Set the result to 0 if it exceeded the time maxProcessingTimesToTrack
		}

		results <- Result{Task: task, Factorial: result}
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano())) // Inicializálja a véletlenszám-generátort

	numTasks := 100

	tasks := make(chan Task, numTasks)
	results := make(chan Result, numTasks)

	var wg sync.WaitGroup

	go GenerateTasks(numTasks, tasks)

	numWorkers := runtime.NumCPU() + 1
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go Worker(i, tasks, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	sortedResult := make([]Result, numTasks)
	for r := range results {
		sortedResult[r.Task.ID] = r
	}

	sort.SliceStable(sortedResult, func(i, j int) bool {
		return sortedResult[i].Task.ID < sortedResult[j].Task.ID
	})

	for _, result := range sortedResult {
		fmt.Printf("%d. task: %d! = %d\n", result.Task.ID, result.Task.Value, result.Factorial)
	}

}

func calculateAverageProcessingTime() time.Duration {
	processingTimesLock.Lock()
	defer processingTimesLock.Unlock()

	var sum time.Duration
	for _, t := range processingTimes {
		sum += t
	}
	if len(processingTimes) == 0 {
		return 0
	}
	return sum / time.Duration(len(processingTimes))
}

// Factorial calculates the factorial of a non-negative integer n
// using the big.Int type to handle large numbers.
func Factorial(n int64) *big.Int {
	if n < 0 {
		return big.NewInt(0) // Returns 0 for negative inputs as factorial is undefined
	}

	result := big.NewInt(1) // Initializes the result as 1, the factorial of 0
	for i := int64(1); i <= n; i++ {
		// Multiplies the result by i for each iteration
		result.Mul(result, big.NewInt(i))
	}

	return result
}
