// Package tasque provides a generic task queue that can be used to process tasks concurrently.
// It supports storing tasks in a storage manager if the queue is full.
package tasque

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrQueueIsFull = errors.New("message queue is full")

// TasksQueue is a generic task queue.
type TasksQueue[T Task] struct {
	channel          chan T
	cacheCapacity    int
	errorHandler     ErrorHandler
	taskStoreManager TaskStorageSaveGetDeleter[T]
	errorHandlerMu   sync.Mutex
}

// NewTasksQueue creates a new TasksQueue with the given cache capacity.
func NewTasksQueue[T Task](cacheCapacity int) *TasksQueue[T] {
	return &TasksQueue[T]{
		channel:       make(chan T, cacheCapacity),
		cacheCapacity: cacheCapacity,
		errorHandler:  func(err error) {},
	}
}

// addBatchFromStorageIfNeeded adds a batch of tasks from storage to the queue if needed.
func (queue *TasksQueue[T]) addBatchFromStorageIfNeeded(ctx context.Context) {
	if queue.taskStoreManager == nil {
		return
	}
	errHandlerFunc := queue.errorHandler

	if len(queue.channel)+queue.taskStoreManager.TaskFromStorageBatchCount() < queue.cacheCapacity {
		items, err := queue.taskStoreManager.GetTasksFromStorage(ctx)
		if err != nil {
			errHandlerFunc(err)
			return
		}
		for _, item := range items {
			errChan := make(chan error)
			go queue.sendToQueue(errChan, item)
			err := <-errChan
			if err != nil {
				errHandlerFunc(err)
				return
			}
			if err = queue.taskStoreManager.DeleteTaskFromStorage(ctx, &item); err != nil {
				errHandlerFunc(err)
			}
		}
	}
}

// doTarget performs the given task.
func (queue *TasksQueue[T]) doTarget(ctx context.Context, target T) {
	queue.errorHandlerMu.Lock()
	handler := queue.errorHandler
	queue.errorHandlerMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			handler(fmt.Errorf("panic occurred: %v", r))
		}
	}()
	target.Do(ctx)
}

// startQueue starts the task queue.
func (queue *TasksQueue[T]) startQueue(ctx context.Context) {
	for {
		queue.addBatchFromStorageIfNeeded(ctx)
		select {
		case <-ctx.Done():
			return
		case target := <-queue.channel:
			queue.doTarget(ctx, target)
		}
	}
}

// StartQueue starts the task queue in a new goroutine.
func (queue *TasksQueue[T]) StartQueue(ctx context.Context) {
	go queue.startQueue(ctx)
}

// sendToQueue sends a task to the queue.
func (queue *TasksQueue[T]) sendToQueue(errChan chan error, item T) {
	select {
	case queue.channel <- item:
		errChan <- nil
		return
	default:
		errChan <- ErrQueueIsFull
		return
	}
}

// SendToQueue sends a task to the queue. If the queue is full and a task store manager is set,
// the task is saved to storage.
func (queue *TasksQueue[T]) SendToQueue(ctx context.Context, item T) (err error) {
	errChan := make(chan error)
	go queue.sendToQueue(errChan, item)
	err = <-errChan
	if err != nil && errors.Is(err, ErrQueueIsFull) && queue.taskStoreManager != nil {
		return queue.taskStoreManager.SaveTaskToStorage(ctx, &item)
	} else if err != nil {
		queue.errorHandler(err)
	}
	return
}

// SetTaskStoreManager sets the task store manager.
func (queue *TasksQueue[T]) SetTaskStoreManager(manager TaskStorageSaveGetDeleter[T]) {
	queue.taskStoreManager = manager
}

// SetErrorHandler sets the error handler.
func (queue *TasksQueue[T]) SetErrorHandler(handler ErrorHandler) {
	queue.errorHandlerMu.Lock()
	defer queue.errorHandlerMu.Unlock()
	queue.errorHandler = handler
}
