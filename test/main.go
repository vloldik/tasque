package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vloldik/tasque"
)

type TaskData struct {
	ID   int
	Data string
}

func main() {
	wg := sync.WaitGroup{}
	// Define a task function
	task := func(ctx context.Context, data TaskData) bool {
		fmt.Printf("Executing task %d with data: %s\n", data.ID, data.Data)
		wg.Done()
		// ... Your task logic goes here ...
		return false
	}

	// Define a task storage manager
	taskStorageManager := &tasque.MemoryTaskStorage[TaskData]{}

	// Define an error handler
	errorHandler := func(err error) (needToReshedule bool) { return true }

	timeout := time.Hour

	// Create a task queue
	taskQueue := tasque.CreateTaskQueue(task, 3, taskStorageManager, errorHandler, timeout)

	// Send tasks to the queue
	for i := 0; i < 10; i++ {
		data := TaskData{
			ID:   i + 1,
			Data: fmt.Sprintf("Data for task %d", i+1),
		}
		err := taskQueue.SendToQueue(context.Background(), data)
		if err != nil {
			log.Printf("Failed to send task to the queue: %v", err)
		}
	}

	wg.Add(10)

	// Start the task queue
	err := taskQueue.StartQueue(context.Background())
	if err != nil {
		log.Printf("Failed to start the task queue: %v", err)
	}

	// Wait for all tasks to complete
	wg.Wait()

	fmt.Println("All tasks completed")
}
