package model

import "math/big"

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
