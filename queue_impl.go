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

type TaskQueueItem[D interface{}, T Task[D]] struct {
	Task T
	Data D
}

// TasksQueue is a generic task queue.
type TasksQueue[D interface{}, T Task[D]] struct {
	channel          chan TaskQueueItem[D, T]
	cacheCapacity    int
	errorHandler     ErrorHandler
	taskStoreManager TaskStorageSaveGetDeleter[D, T]
	errorHandlerMu   sync.Mutex
}

// NewTasksQueue creates a new TasksQueue with the given cache capacity.
func NewTasksQueue[D interface{}, T Task[D]](cacheCapacity int) *TasksQueue[D, T] {
	return &TasksQueue[D, T]{
		channel:       make(chan TaskQueueItem[D, T], cacheCapacity),
		cacheCapacity: cacheCapacity,
		errorHandler:  func(err error) {},
	}
}

// addBatchFromStorageIfNeeded adds a batch of tasks from storage to the queue if needed.
func (queue *TasksQueue[D, T]) addBatchFromStorageIfNeeded(ctx context.Context) {
	if queue.taskStoreManager == nil {
		return
	}
	errHandlerFunc := queue.errorHandler

	if len(queue.channel)+queue.taskStoreManager.TaskFromStorageBatchCount() <= queue.cacheCapacity {
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
func (queue *TasksQueue[D, T]) doTarget(ctx context.Context, target T, data D) {
	queue.errorHandlerMu.Lock()
	handler := queue.errorHandler
	queue.errorHandlerMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			handler(fmt.Errorf("panic occurred: %v", r))
		}
	}()
	target(ctx, data)
}

// startQueue starts the task queue.
func (queue *TasksQueue[D, T]) startQueue(ctx context.Context) {
	for {
		queue.addBatchFromStorageIfNeeded(ctx)
		select {
		case <-ctx.Done():
			return
		case target := <-queue.channel:
			queue.doTarget(ctx, target.Task, target.Data)
		default:

		}
	}
}

// StartQueue starts the task queue in a new goroutine.
func (queue *TasksQueue[D, T]) StartQueue(ctx context.Context) {
	go queue.startQueue(ctx)
}

// sendToQueue sends a task to the queue.
func (queue *TasksQueue[D, T]) sendToQueue(errChan chan error, item TaskQueueItem[D, T]) {
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
func (queue *TasksQueue[D, T]) SendToQueue(ctx context.Context, item TaskQueueItem[D, T]) (err error) {
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
func (queue *TasksQueue[D, T]) SetTaskStoreManager(manager TaskStorageSaveGetDeleter[D, T]) {
	queue.taskStoreManager = manager
}

// SetErrorHandler sets the error handler.
func (queue *TasksQueue[D, T]) SetErrorHandler(handler ErrorHandler) {
	queue.errorHandlerMu.Lock()
	defer queue.errorHandlerMu.Unlock()
	queue.errorHandler = handler
}
