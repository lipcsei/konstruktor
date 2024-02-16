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
	// WorkerID identifies the worker that completed processing the task.
	WorkerID int
}

type Worker struct {
	ID int
	// tasks is a channel from which the worker receives tasks to process.
	tasks <-chan Task
	// results is a channel to which the worker sends processed tasks.
	results chan<- Result
	// quit is a channel used to signal the worker to gracefully shut down.
	quit <-chan struct{}
	// wg is used to signal when the worker has finished processing.
	wg *sync.WaitGroup
	// processingTimes stores the processing times of recent tasks.
	processingTimes []time.Duration
	// maxProcessingTimesToTrack is the maximum number of processing times to consider for calculating the average.
	maxProcessingTimesToTrack int
}

// New initializes and returns a new Worker instance.
func New(id int, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup, quit <-chan struct{}) *Worker {
	return &Worker{
		ID:                        id,
		tasks:                     tasks,
		results:                   results,
		quit:                      quit,
		wg:                        wg,
		maxProcessingTimesToTrack: maxProcessingTimesToTrack,
	}
}

// Start is the main method of the Worker, where it begins processing tasks from the tasks channel.
// It listens for tasks to process and quit signals for shutdown, utilizing a select statement to handle
// both concurrently. If a quit signal is received, the worker stops processing and exits.
func (w *Worker) Start() {
	defer w.wg.Done()

	for {
		select {
		// Attempt to receive a task from the tasks channel.
		case task, ok := <-w.tasks:
			if !ok {
				// If the tasks channel is closed, exit the loop and end the goroutine.
				return
			}

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

			// Update the processingTimes slice.
			w.updateProcessingTimes(processingTime)

			// Calculate the allowed time threshold as 10% above the average time
			allowedTimeThreshold := averageTime + (averageTime / 10)

			// Check if the processing time exceeds the allowed time threshold
			if processingTime > 0 && averageTime > 0 && processingTime > allowedTimeThreshold {
				result = big.NewInt(0) // Override the factorial result with 0.
			}

			// Send the result (either the calculated factorial or 0) to the results channel.
			w.results <- Result{Task: task, Factorial: result}
		case <-w.quit:
			// If a quit signal is received, exit the loop and end the goroutine.
			return
		}
	}
}

// updateProcessingTimes updates the slice of processing times with the latest task processing time.
// It ensures that the slice does not exceed the maximum number of processing times to track.
// Older processing times are removed to maintain the size limit.
func (w *Worker) updateProcessingTimes(processingTime time.Duration) {

	// Check if the processing times slice has reached its maximum capacity.
	if len(w.processingTimes) >= w.maxProcessingTimesToTrack {
		// Remove the oldest processing time to make room for the new one.
		w.processingTimes = w.processingTimes[1:]
	}

	// Add the new processing time to the end of the slice.
	w.processingTimes = append(w.processingTimes, processingTime)
}

// calculateAverageProcessingTime computes the average processing time of the most recent tasks,
// up to the number specified by maxProcessingTimesToTrack.
// It locks the processingTimes slice during calculation to ensure thread-safe access.
// Returns 0 if there are no recorded processing times.
func (w *Worker) calculateAverageProcessingTime() time.Duration {

	var sum time.Duration
	// Sum up all recorded processing times.
	for _, t := range w.processingTimes {
		sum += t
	}

	// Avoid division by zero if no processing times are recorded.
	if len(w.processingTimes) == 0 {
		return 0
	}

	// Calculate and return the average processing time.
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
