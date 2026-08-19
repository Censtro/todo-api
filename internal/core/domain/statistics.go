package domain

import "time"

type Statistics struct {
	TasksCreated               int
	TasksCompleted             int
	TasksCompletedRate         *float64
	TasksAverageCompletionTime *time.Duration
}

func NewStatistics(
	tasksCreated int,
	taskCompleted int,
	tasksCompletedRate *float64,
	taskAverageCompletionTime *time.Duration,
) Statistics {
	return Statistics{
		TasksCreated:               taskCompleted,
		TasksCompleted:             taskCompleted,
		TasksCompletedRate:         tasksCompletedRate,
		TasksAverageCompletionTime: taskAverageCompletionTime,
	}
}
