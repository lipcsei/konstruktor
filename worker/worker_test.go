package worker

import (
	"github.com/lipcsei/konstruktor/model"
	"math/big"
	"sync"
	"testing"
	"time"
)

func TestWorker_Start_TaskProcessingTimeLimit(t *testing.T) {
	// Setup tasks with a single task value. In this case, the task will be forced to exceed processing time limits.
	tasks := []int64{3}
	expectedResults := []*big.Int{
		big.NewInt(0), // Expecting a result of 0 due to processing time limit exceeded.
	}

	// Simulate a delay in task processing to trigger the processing time limit.
	simulateDelay = func() {
		time.Sleep(500 * time.Millisecond)
	}
	defer func() { simulateDelay = nil }()

	taskChannel := make(chan model.Task, len(tasks))
	resultChannel := make(chan model.Result, len(tasks))
	quit := make(chan struct{})

	var wg sync.WaitGroup
	testWorker := New(1, taskChannel, resultChannel, &wg, quit)

	testWorker.maxProcessingTimesToTrack = 3
	processingTimes = []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
	}

	wg.Add(1)
	go testWorker.Start()

	for i, task := range tasks {
		taskChannel <- model.Task{ID: i, Value: task}
	}
	close(taskChannel)

	go func() {
		wg.Wait()
		close(resultChannel)
		close(quit)
	}()

	for i, expectedResult := range expectedResults {
		result := <-resultChannel
		if result.Factorial.Cmp(expectedResult) != 0 {
			t.Errorf("Task %d expected result %v, got %v", tasks[i], expectedResult, result.Factorial)
		}
	}
}

func TestWorker_Start_TaskProcessingOrder(t *testing.T) {

	tasks := []int64{3, 5, 7}
	expectedResults := []*big.Int{
		big.NewInt(6),    // 3!
		big.NewInt(120),  // 5!
		big.NewInt(5040), // 7!
	}

	taskChannel := make(chan model.Task, len(tasks))
	resultChannel := make(chan model.Result, len(tasks))
	quit := make(chan struct{})

	var wg sync.WaitGroup
	testWorker := New(1, taskChannel, resultChannel, &wg, quit)
	// override
	testWorker.maxProcessingTimesToTrack = 5

	wg.Add(1)
	go testWorker.Start()

	for i, task := range tasks {
		taskChannel <- model.Task{ID: i, Value: task}
	}
	close(taskChannel)

	go func() {
		wg.Wait()
		close(resultChannel)
		close(quit)
	}()

	sortedResults := SortResults(resultChannel, len(tasks))

	for i, expectedResult := range expectedResults {
		if sortedResults[i].Factorial.Cmp(expectedResult) != 0 {
			t.Errorf("Task %d expected result %v, got %v", tasks[i], expectedResult, sortedResults[i].Factorial)
		}
	}
}

// TestCalculateAverageProcessingTime tests the calculateAverageProcessingTime function to ensure
// it correctly calculates the average processing time from a set of durations.
func TestCalculateAverageProcessingTime(t *testing.T) {
	testWorker := New(1, nil, nil, nil, nil)

	// Setup: Clear and then set predefined processing times for testing
	processingTimes = []time.Duration{} // Clear existing processing times
	testDurations := []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
	}
	processingTimes = append(processingTimes, testDurations...)

	// Expected average calculation
	var expectedSum time.Duration
	for _, d := range testDurations {
		expectedSum += d
	}
	expectedAverage := expectedSum / time.Duration(len(testDurations))

	// Test: Calculate the average processing time
	average := testWorker.calculateAverageProcessingTime()

	// Assert: Check if the calculated average is as expected
	if average != expectedAverage {
		t.Errorf("calculateAverageProcessingTime() = %v, want %v", average, expectedAverage)
	}
}
