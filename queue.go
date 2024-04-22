// Package tasque provides a generic task queue and related interfaces.
// It allows tasks to be processed concurrently and supports error handling and storage management.
package tasque

import "context"

// Task is the interface that wraps the basic Do method.
//
// Do is called to perform the task.
type Task[D interface{}] func(ctx context.Context, data D) (needToReshedule bool)

// ErrorHandler is a function that handles errors.
type ErrorHandler func(err error) (needToReshedule bool)

// ErrorHandlerSetter is the interface that wraps the SetErrorHandler method.
//
// SetErrorHandler sets the error handler for handling errors.
type ErrorHandlerSetter interface {
	SetErrorHandler(handler ErrorHandler)
}

// TaskStorageSaveGetDeleter is the interface that wraps methods for saving, getting, and deleting tasks from storage.
//
// SaveTaskToStorage saves a task to storage.
// GetTasksFromStorage retrieves tasks from storage.
// DeleteTaskFromStorage deletes a task from storage.
// TaskFromStorageBatchCount returns the number of tasks to retrieve in a single batch.
type TaskStorageSaveGetDeleter[D interface{}] interface {
	SaveTaskToStorage(ctx context.Context, data *D) error
	GetTasksFromStorage(ctx context.Context, count int) ([]D, error)
	DeleteTaskFromStorage(ctx context.Context, data *D) error
}

// TasksQueueManger is the interface that wraps the StartQueue and SendToQueue methods.
//
// StartQueue starts the task queue.
// SendToQueue sends a task to the queue.
type TasksQueueManger[D interface{}] interface {
	StartQueue(ctx context.Context)
	SendToQueue(ctx context.Context, data D) error
}

// TasksQueueManagerWithStoreAndErrorHandler is the interface that combines TasksQueueManger and TaskStorageSaveGetDeleter,
// and includes the ErrorHandlerSetter for setting the error handler.
//
// It inherits the methods from TasksQueueManger and TaskStorageSaveGetDeleter, and adds the SetErrorHandler method from ErrorHandlerSetter.
type TasksQueueManagerWithStoreAndErrorHandler[D interface{}] interface {
	TasksQueueManger[D]
	SetTaskStoreManager(manager TaskStorageSaveGetDeleter[D])
	ErrorHandlerSetter
}
