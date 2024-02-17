package generator

import (
	"github.com/lipcsei/konstruktor/model"
	"testing"
)

func TestGenerateTasks(t *testing.T) {
	numTasks := 100 //
	tasksChan := make(chan model.Task, numTasks)

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
