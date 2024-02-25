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
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vloldik/tasque"
	"github.com/vloldik/tasquelite"
)

type MailerContextKey string

var MailerContextValueKey = MailerContextKey("mailer")

// Type for e-mail task
type EmailTask struct {
	ID        int `gorm:"primaryKey"`
	Recipient string
	Subject   string
	Body      string
}

type Sender struct {
	wg *sync.WaitGroup
}

func (sender Sender) Send(value int) {
	time.Sleep(time.Second)
	fmt.Printf("sending value %d\n", value)
	sender.wg.Done()
}

// Do sends the email. This is a placeholder for the actual email sending logic.
func (et EmailTask) Do(ctx context.Context) {
	mailer := ctx.Value(MailerContextValueKey)
	if sender, ok := mailer.(*Sender); ok {
		sender.Send(et.ID)
	} else {
		panic(errors.New("invalid mailer"))
	}
}

func main() {
	// Create a new context with a background.
	ctx := context.Background()

	// Create a new WaitGroup to wait for all tasks to complete.
	wg := &sync.WaitGroup{}

	// Create a new task queue with a capacity of 5.
	queue := tasque.NewTasksQueue[EmailTask](5)

	// Create a new sender with the WaitGroup.
	sender := Sender{wg: wg}

	// Add the sender to the context.
	ctx = context.WithValue(ctx, MailerContextValueKey, &sender)

	// Create a new GORM task storage manager for EmailTask.
	storage, err := tasquelite.NewGormTaskStorageManager[EmailTask]("test.db", &EmailTask{}, 2)
	queue.SetTaskStoreManager(storage)

	if err != nil {
		// Handle the error if storage creation fails.
		panic(err)
	}

	// Define a target email task.
	target := EmailTask{
		Recipient: "foo@bar.com",
		Subject:   "<div>Hello!</div>",
		Body:      "Some important Subj",
	}

	// Fill the queue with 5 tasks.
	for i := 0; i < 5; i++ {
		// Send the task to the queue.
		// ID will be 0
		queue.SendToQueue(ctx, target)
		// Increment the WaitGroup counter for each task.
		sender.wg.Add(1)
	}

	// Try to send 10 tasks to the queue, which will be stored in memory.
	for i := 0; i < 10; i++ {
		// Send the task to the queue.
		// ID from storage
		queue.SendToQueue(ctx, target)
		// Increment the WaitGroup counter for each task.
		sender.wg.Add(1)
	}
	queue.StartQueue(ctx)
	// Wait for all tasks to complete.
	wg.Wait()
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
