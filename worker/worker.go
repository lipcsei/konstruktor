package worker

import (
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const maxProcessingTimesToTrack = 20

// Global variables for storing processing times and for synchronization
var simulateDelay func()

// Task represents a unit of work to process.
type Task struct {
	// ID is an unique identifier that also determines result sorting order
	ID int
	// Value specifies the number for which the calcFactorial is to be calculated.
	Value int64
}

// Result represents the outcome of a Task, including its factorial result
type Result struct {
	Task      Task
	Factorial *big.Int
}

type Worker struct {
	ID      int
	tasks   <-chan Task
	results chan<- Result
	wg      *sync.WaitGroup

	// processingTimes is a slice to store the processing times of tasks
	processingTimes []time.Duration
	// processingTimesLock is a mutex to synchronize access to the processingTimes slice
	processingTimesLock sync.Mutex
	// Maximum number of processing times to keep track of for average calculation
	maxProcessingTimesToTrack int
}

func New(id int, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) Worker {
	return Worker{
		ID:                        id,
		tasks:                     tasks,
		results:                   results,
		wg:                        wg,
		maxProcessingTimesToTrack: maxProcessingTimesToTrack,
	}
}

func (w *Worker) Start() {
	defer w.wg.Done()
	for task := range w.tasks {
		startTime := time.Now()

		if simulateDelay != nil {
			simulateDelay()
		}

		result := calcFactorial(task.Value)

		processingTime := time.Since(startTime)

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
func GenerateTasks(numTasks int, tasks chan<- Task) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
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
