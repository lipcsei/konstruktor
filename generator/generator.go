package generator

import (
	"github.com/lipcsei/konstruktor/model"
	"math/rand"
	"time"
)

// GenerateTasks generates a specified number of tasks with random values and sends them on a channel
// Each task's value is randomly chosen between 3 and 1000, inclusive.
func GenerateTasks(numTasks int, tasks chan<- model.Task) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < numTasks; i++ {
		// Generate a random number between 3 and 1000
		randomNumber := int64(rand.Intn(998) + 3)

		// Create a task
		task := model.Task{
			ID:    i,
			Value: randomNumber,
		}
		// Send the new task
		tasks <- task
	}

	// Signal to processors that there are no more tasks
	close(tasks)
}
