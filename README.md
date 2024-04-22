# Tasque

Tasque is a Go package that provides a task queue implementation with thread management. It allows you to create a task queue with a specified number of threads and provides methods to send tasks to the queue, start the queue, and manage task storage.

## Installation

To use Tasque in your Go project, you can simply import the package:

```go
import "github.com/your_username/tasque"
```

Replace `your_username` with your actual GitHub username or the name of your project where you want to import Tasque.

## Usage

### Creating a Task Queue

To create a task queue, use the `CreateTaskQueue` function:

```go
taskQueue := tasque.CreateTaskQueue(task, threadCount, taskStorageManager, errorHandler, taskTimeoutDuration)
```

- `task`: A function or method that represents the task to be executed. It should have the signature `func(ctx context.Context, data D) bool`, where `D` is the type of data for the task.
- `threadCount`: The number of threads to use for executing tasks concurrently.
- `taskStorageManager`: An implementation of the `TaskStorageSaveGetDeleter` interface that handles saving, retrieving, and deleting tasks from storage.
- `errorHandler`: An implementation of the `ErrorHandler` interface that handles errors occurred during task execution.

### Sending Tasks to the Queue

To send a task to the queue, use the `SendToQueue` method:

```go
err := taskQueue.SendToQueue(ctx, data)
```

- `ctx`: The context.Context for the task execution.
- `data`: The data for the task.

### Starting the Queue

To start the task queue, use the `StartQueue` method:

```go
err := taskQueue.StartQueue(ctx)
```

- `ctx`: The context.Context for the task queue.

### Task Storage

The `taskStorageManager` parameter in the `CreateTaskQueue` function requires an implementation of the `TaskStorageSaveGetDeleter` interface:

```go
type TaskStorageSaveGetDeleter[D struct{}] interface {
 SaveTaskToStorage(ctx context.Context, data *D) error
 GetTasksFromStorage(ctx context.Context, count int) ([]D, error)
 DeleteTaskFromStorage(ctx context.Context, data *D) error
}
```

You can implement this interface to provide your own storage mechanism for tasks.

### Error Handling

The `errorHandler` parameter in the `CreateTaskQueue` function requires an implementation of the `ErrorHandler` interface:

```go
type ErrorHandler interface {
 HandleError(err error) bool
}
```

You can implement this interface to define how errors should be handled during task execution.

## Example

Here's an example usage of Tasque:

```go
func main() {
	wg := sync.WaitGroup{}
	// Define a task function
	task := func(ctx context.Context, data TaskData) bool {
		fmt.Printf("Executing task %d with data: %s\n", data.ID, data.Data)
		wg.Done()
		// ... Your task logic goes here ...

		// Return if the task need to re-shedule
		return false
	}

	// Define a task storage manager
	taskStorageManager := &tasque.MemoryTaskStorage[TaskData]{}

	// Define an error handler
	errorHandler := func(err error) (needToReshedule bool) { return true // Return if the task need to re-shedule on error }

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


```

In this example, we define a task function that prints the task ID and data when executed. We use the `MemoryTaskStorage` as the task storage manager and the `DefaultErrorHandler` for error handling. We then send 10 tasks to the queue, start the queue, and wait for all tasks to complete.

## Contributing

Contributions to Tasque are welcome! 
