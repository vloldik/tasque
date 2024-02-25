# Tasque: A Generic Task Queue Package

[![GoDoc](https://godoc.org/github.com/vloldik/tasque?status.svg)](https://godoc.org/github.com/vloldik/tasque)
[![Go Report Card](https://goreportcard.com/badge/github.com/vloldik/tasque)](https://goreportcard.com/report/github.com/vloldik/tasque)

Tasque is a Go package that provides a generic task queue for concurrent processing of tasks. It supports storing tasks in a storage manager if the queue is full, and allows custom error handling.

## Features

- **Generic Task Queue**: Process tasks concurrently using a generic task queue.
- **Storage Management**: Store tasks in a storage manager if the queue is full.
- **Custom Error Handling**: Set a custom error handler to handle errors during task processing.

## Installation

To install Tasque, use `go get`:

```sh
go get github.com/vloldik/tasque
```
**Usage**
Here's a basic example of how to use Tasque:
```go
package main

import (
	"context"
	"fmt"
	"github.com/vloldik/tasque"
)

// SimpleTask is a simple task that prints a message and accesses context data.
type SimpleTask struct {
	Message string
}

// Do performs the task.
func (t *SimpleTask) Do(ctx context.Context) {
	// Access context data.
	value := ctx.Value("key").(string)
	fmt.Printf("Context value: %s\n", value)

	// Perform the task.
	fmt.Println(t.Message)
}

func main() {
	// Create a new task queue with a cache capacity of 10.
	queue := tasque.NewTasksQueue[*SimpleTask](10)

	// Set a custom error handler.
	queue.SetErrorHandler(func(err error) {
		fmt.Printf("Error occurred: %v\n", err)
	})

	// Start the task queue.
	ctx := context.Background()
	ctx = context.WithValue(ctx, "key", "value") // Add context data.
	queue.StartQueue(ctx)

	// Send a task to the queue.
	task := &SimpleTask{Message: "Hello, Tasque!"}
	if err := queue.SendToQueue(ctx, task); err != nil {
		fmt.Printf("Failed to send task to queue: %v\n", err)
	}

	// Wait for the task to be processed.
	select {}
}
```

## Store implementations
* SQLite, gorm [tasquelite](https://github.com/vloldik/tasquelite)

## Contributing
Contributions are welcome! If you find a bug or have a feature request, please open an issue. If you would like to contribute code, please follow these steps:
1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Write tests for your changes.
4. Make your changes and ensure that all tests pass.
5. Submit a pull request with a detailed description of your changes.

License
Tasque is released under the MIT License.
