package worker

import (
	"github.com/lipcsei/konstruktor/model"
	"github.com/lipcsei/konstruktor/utils"
	"math/big"
	"sort"
	"sync"
	"time"
)

// processingTimes stores the processing times of recent tasks.
var processingTimes []time.Duration

// processingTimesLock synchronizes access to the processingTimes slice.
var processingTimeLock sync.Mutex

// maxProcessingTimesToTrack specifies the length of the slice that stores processing times of tasks.
// This is used to calculate the average processing time by keeping a limited history of recent processing times.
const maxProcessingTimesToTrack = 20

// simulateDelay is a global variable that allows for simulating a delay in task processing.
// It can be set to a function that pauses execution, typically used for testing.
var simulateDelay func()

type Worker struct {
	ID int
	// tasks is a channel from which the worker receives tasks to process.
	tasks <-chan model.Task
	// results is a channel to which the worker sends processed tasks.
	results chan<- model.Result
	// quit is a channel used to signal the worker to gracefully shut down.
	quit <-chan struct{}
	// wg is used to signal when the worker has finished processing.
	wg *sync.WaitGroup

	// maxProcessingTimesToTrack is the maximum number of processing times to consider for calculating the average.
	maxProcessingTimesToTrack int
}

// New initializes and returns a new Worker instance.
func New(id int, tasks <-chan model.Task, results chan<- model.Result, wg *sync.WaitGroup, quit <-chan struct{}) *Worker {
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
			result := utils.CalcFactorial(task.Value)

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
			w.results <- model.Result{Task: task, Factorial: result, WorkerID: w.ID}
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
	processingTimeLock.Lock()
	defer processingTimeLock.Unlock()
	// Check if the processing times slice has reached its maximum capacity.
	if len(processingTimes) >= w.maxProcessingTimesToTrack {
		// Remove the oldest processing time to make room for the new one.
		processingTimes = processingTimes[1:]
	}

	// Add the new processing time to the end of the slice.
	processingTimes = append(processingTimes, processingTime)
}

// calculateAverageProcessingTime computes the average processing time of the most recent tasks,
// up to the number specified by maxProcessingTimesToTrack.
// It locks the processingTimes slice during calculation to ensure thread-safe access.
// Returns 0 if there are no recorded processing times.
func (w *Worker) calculateAverageProcessingTime() time.Duration {
	processingTimeLock.Lock()
	defer processingTimeLock.Unlock()
	var sum time.Duration
	// Sum up all recorded processing times.
	for _, t := range processingTimes {
		sum += t
	}

	// Avoid division by zero if no processing times are recorded.
	if len(processingTimes) == 0 {
		return 0
	}

	// Calculate and return the average processing time.
	return sum / time.Duration(len(processingTimes))
}

// SortResults sorts the results based on their task ID and returns a slice of sorted results.
func SortResults(results chan model.Result, length int) []model.Result {
	sortedResult := make([]model.Result, length)
	for r := range results {
		sortedResult[r.Task.ID] = r
	}

	sort.SliceStable(sortedResult, func(i, j int) bool {
		return sortedResult[i].Task.ID < sortedResult[j].Task.ID
	})
	return sortedResult
}
