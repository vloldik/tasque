# Tasque: A Generic Task Queue Package

[![GoDoc](https://godoc.org/github.com/vloldik/tasque?status.svg)](https://godoc.org/github.com/vloldik/tasque)
[![Go Report Card](https://goreportcard.com/badge/github.com/vloldik/tasque)](https://goreportcard.com/report/github.com/vloldik/tasque)

Tasque is a Go package that provides a generic task queue for concurrent processing of tasks. It supports storing tasks in a storage manager if the queue is full, and allows custom error handling.

## Features

- **Generic Task Queue**: Process tasks using a generic task queue.
- **Storage Management**: Store tasks in a storage manager if the queue is full.
- **Custom Error Handling**: Set a custom error handler to handle errors during task processing.

## Installation

To install Tasque, use `go get`:

```sh
go get github.com/vloldik/tasque
```

## Usecase

```go
ctx := context.Background()
countDown := sync.WaitGroup{}
countDown.Add(2)

maxParallelJobs := 10
maxQueueSizeInRam := 2

queue := NewTasksQueue(func(ctx context.Context, data int) {
	fmt.Printf("Hello from queue! Item №%d\n", data)
	time.Sleep(time.Second)
	countDown.Done()
}, maxQueueSizeInRam, maxParallelJobs)

queue.SetTaskStoreManager(&yourmanager)
queue.SetErrorHandler(func(err error) { fmt.Printf("An error occurred %e\n", err) })
queue.SendToQueue(ctx, 1)
queue.SendToQueue(ctx, 2)
queue.StartQueue(ctx)
countDown.Wait()
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
