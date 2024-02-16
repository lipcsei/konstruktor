package worker

import (
	"fmt"
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

	taskChannel := make(chan Task, len(tasks))
	resultChannel := make(chan Result, len(tasks))
	quit := make(chan struct{})

	var wg sync.WaitGroup
	testWorker := New(1, taskChannel, resultChannel, &wg, quit)

	testWorker.maxProcessingTimesToTrack = 3
	testWorker.processingTimes = []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
	}

	wg.Add(1)
	go testWorker.Start()

	for i, task := range tasks {
		taskChannel <- Task{ID: i, Value: task}
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

	taskChannel := make(chan Task, len(tasks))
	resultChannel := make(chan Result, len(tasks))
	quit := make(chan struct{})

	var wg sync.WaitGroup
	testWorker := New(1, taskChannel, resultChannel, &wg, quit)
	// override
	testWorker.maxProcessingTimesToTrack = 5

	wg.Add(1)
	go testWorker.Start()

	for i, task := range tasks {
		taskChannel <- Task{ID: i, Value: task}
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

func TestCalcFactorial(t *testing.T) {
	tests := []struct {
		name     string
		n        int64
		expected string
	}{
		{"0!", -1, "0"},
		{"0!", 0, "1"},
		{"1!", 1, "1"},
		{"5!", 5, "120"},
		{"10!", 10, "3628800"},
		{"40!", 40, "815915283247897734345611269596115894272000000000"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.name), func(t *testing.T) {
			result := calcFactorial(test.n)
			if result.String() != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result.String())
			}
		})
	}
}

// TestCalculateAverageProcessingTime tests the calculateAverageProcessingTime function to ensure
// it correctly calculates the average processing time from a set of durations.
func TestCalculateAverageProcessingTime(t *testing.T) {
	testWorker := New(1, nil, nil, nil, nil)

	// Setup: Clear and then set predefined processing times for testing
	testWorker.processingTimes = []time.Duration{} // Clear existing processing times
	testDurations := []time.Duration{
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
	}
	testWorker.processingTimes = append(testWorker.processingTimes, testDurations...)

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

func TestGenerateTasks(t *testing.T) {
	numTasks := 100 //
	tasksChan := make(chan Task, numTasks)

	GenerateTasks(numTasks, tasksChan)

	generatedTasks := 0
	for task := range tasksChan {
		generatedTasks++
		if task.Value < 3 || task.Value > 1000 {
			t.Errorf("Task value out of expected range: got %v, want between 3 and 1000", task.Value)
		}
	}

	if generatedTasks != numTasks {
		t.Errorf("Incorrect number of tasks generated: got %v, want %v", generatedTasks, numTasks)
	}
}
