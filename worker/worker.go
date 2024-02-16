package worker

import (
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// maxProcessingTimesToTrack specifies the length of the slice that stores processing times of tasks.
// This is used to calculate the average processing time by keeping a limited history of recent processing times.
const maxProcessingTimesToTrack = 20

// simulateDelay is a global variable that allows for simulating a delay in task processing.
// It can be set to a function that pauses execution, typically used for testing.
var simulateDelay func()

// Task represents a unit of work to process.
// It contains a unique identifier and a value for which the factorial will be calculated.
type Task struct {
	// ID is an unique identifier that also determines result sorting order.
	ID int
	// Value specifies the number for which the factorial is to be calculated.
	Value int64
}

// Result represents the outcome of processing a Task, including its factorial result.
type Result struct {
	// Task is the original task that was processed.
	Task Task
	// Factorial is the calculated factorial of the task's value.
	Factorial *big.Int
}

type Worker struct {
	ID int
	// tasks is a channel from which the worker receives tasks to process.
	tasks <-chan Task
	// results is a channel to which the worker sends processed tasks.
	results chan<- Result
	// wg is used to signal when the worker has finished processing.
	wg *sync.WaitGroup
	// processingTimes stores the processing times of recent tasks.
	processingTimes []time.Duration
	// processingTimesLock synchronizes access to the processingTimes slice.
	processingTimesLock sync.Mutex
	// maxProcessingTimesToTrack is the maximum number of processing times to consider for calculating the average.
	maxProcessingTimesToTrack int
}

// New initializes and returns a new Worker instance.
func New(id int, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) Worker {
	return Worker{
		ID:                        id,
		tasks:                     tasks,
		results:                   results,
		wg:                        wg,
		maxProcessingTimesToTrack: maxProcessingTimesToTrack,
	}
}

// Start begins processing tasks from the tasks channel.
// For each task, it calculates the factorial, taking into account a possible delay and processing time limits.
func (w *Worker) Start() {
	defer w.wg.Done()
	for task := range w.tasks {
		// Record the start time of the task processing to measure its duration.
		startTime := time.Now()

		if simulateDelay != nil {
			// If a delay function is defined, invoke it. Useful for testing.
			simulateDelay()
		}

		// Calculate the factorial of the task's value.
		result := calcFactorial(task.Value)

		// Determine the total processing time for the task.
		processingTime := time.Since(startTime)

		// Calculate the current average processing time of recent tasks.
		averageTime := w.calculateAverageProcessingTime()

		w.processingTimesLock.Lock()
		if len(w.processingTimes) >= w.maxProcessingTimesToTrack {
			w.processingTimes = w.processingTimes[1:] // Az első elem eltávolítása
		}
		w.processingTimes = append(w.processingTimes, processingTime)
		w.processingTimesLock.Unlock()

		// Calculate the allowed time threshold as 10% above the average time
		allowedTimeThreshold := averageTime + (averageTime / 10)

		// Check if the processing time exceeds the allowed time threshold
		if processingTime > 0 && averageTime > 0 && processingTime > allowedTimeThreshold {
			result = big.NewInt(0) // Set the result to 0 if it exceeded the time maxProcessingTimesToTrack
		}

		// Send the result (either the calculated factorial or 0) to the results channel.
		w.results <- Result{Task: task, Factorial: result}
	}
}

func (w *Worker) calculateAverageProcessingTime() time.Duration {
	w.processingTimesLock.Lock()
	defer w.processingTimesLock.Unlock()

	var sum time.Duration
	for _, t := range w.processingTimes {
		sum += t
	}
	if len(w.processingTimes) == 0 {
		return 0
	}
	return sum / time.Duration(len(w.processingTimes))
}

// SortResults sorts the results based on their task ID and returns a slice of sorted results.
func SortResults(results chan Result, length int) []Result {
	sortedResult := make([]Result, length)
	for r := range results {
		sortedResult[r.Task.ID] = r
	}

	sort.SliceStable(sortedResult, func(i, j int) bool {
		return sortedResult[i].Task.ID < sortedResult[j].Task.ID
	})
	return sortedResult
}

// GenerateTasks generates a specified number of tasks with random values and sends them on a channel
// Each task's value is randomly chosen between 3 and 1000, inclusive.
func GenerateTasks(numTasks int, tasks chan<- Task) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < numTasks; i++ {
		// Generate a random number between 3 and 1000
		randomNumber := int64(rand.Intn(998) + 3)

		// Create a task
		task := Task{
			ID:    i,
			Value: randomNumber,
		}
		// Send the new task
		tasks <- task
	}

	// Signal to processors that there are no more tasks
	close(tasks)
}

// factorial calculates the factorial of a non-negative integer n
// using the big.Int type to handle large numbers.
func calcFactorial(n int64) *big.Int {
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
