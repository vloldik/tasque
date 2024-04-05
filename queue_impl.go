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
type TasksQueue[D interface{}] struct {
	target           Task[D]
	dataChannel      chan D
	cacheCapacity    int
	errorHandler     ErrorHandler
	taskStoreManager TaskStorageSaveGetDeleter[D]
	errorHandlerMu   sync.Mutex
	maxParallelJobs  int
	wg               sync.WaitGroup
}

// NewTasksQueue creates a new TasksQueue with the given cache capacity.
func NewTasksQueue[D interface{}](function Task[D], cacheCapacity int, maxParallelJobs int) *TasksQueue[D] {
	return &TasksQueue[D]{
		dataChannel:     make(chan D, cacheCapacity),
		cacheCapacity:   cacheCapacity,
		errorHandler:    func(err error) {},
		target:          function,
		maxParallelJobs: maxParallelJobs,
	}
}

func (queue *TasksQueue[D]) SetFunction(function Task[D]) {
	queue.target = function
}

// addBatchFromStorageIfNeeded adds a batch of tasks from storage to the queue if needed.
func (queue *TasksQueue[D]) addBatchFromStorageIfNeeded(ctx context.Context) {
	if queue.taskStoreManager == nil {
		return
	}
	errHandlerFunc := queue.errorHandler

	if len(queue.dataChannel)+queue.taskStoreManager.TaskFromStorageBatchCount() <= queue.cacheCapacity {
		items, err := queue.taskStoreManager.GetTasksFromStorage(ctx)
		if err != nil {
			errHandlerFunc(err)
			return
		}
		for _, item := range items {
			queue.SendToQueue(ctx, item)
			if err = queue.taskStoreManager.DeleteTaskFromStorage(ctx, &item); err != nil {
				errHandlerFunc(err)
			}
		}
	}
}

// doTarget performs the given task.
func (queue *TasksQueue[D]) doTarget(ctx context.Context, data D) {
	queue.errorHandlerMu.Lock()
	handler := queue.errorHandler
	queue.errorHandlerMu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			handler(fmt.Errorf("panic occurred: %v", r))
		}
	}()
	queue.target(ctx, data)
}

// startQueue starts the task queue.
func (queue *TasksQueue[D]) startQueue(ctx context.Context) {
	semaphore := make(chan struct{}, queue.maxParallelJobs)

	for {
		queue.addBatchFromStorageIfNeeded(ctx)
		select {
		case <-ctx.Done():
			queue.wg.Wait() // wait for all tasks to finish
			return
		case data := <-queue.dataChannel:
			queue.wg.Add(1)
			semaphore <- struct{}{} // acquire a slot
			go func(data D) {
				defer func() {
					<-semaphore // release the slot
					queue.wg.Done()
				}()
				queue.doTarget(ctx, data)
			}(data)
		default:

		}
	}
}

// StartQueue starts the task queue in a new goroutine.
func (queue *TasksQueue[D]) StartQueue(ctx context.Context) {
	go queue.startQueue(ctx)
}

// sendToQueue sends a task to the queue.
func (queue *TasksQueue[D]) sendToQueue(errChan chan error, data D) {
	select {
	case queue.dataChannel <- data:
		errChan <- nil
		return
	default:
		errChan <- ErrQueueIsFull
		return
	}
}

// SendToQueue sends a task to the queue. If the queue is full and a task store manager is set,
// the task is saved to storage.
func (queue *TasksQueue[D]) SendToQueue(ctx context.Context, data D) (err error) {
	errChan := make(chan error)
	go queue.sendToQueue(errChan, data)
	err = <-errChan
	if err != nil && errors.Is(err, ErrQueueIsFull) && queue.taskStoreManager != nil {
		return queue.taskStoreManager.SaveTaskToStorage(ctx, &data)
	} else if err != nil {
		queue.errorHandler(err)
	}
	return
}

// SetTaskStoreManager sets the task store manager.
func (queue *TasksQueue[D]) SetTaskStoreManager(manager TaskStorageSaveGetDeleter[D]) {
	queue.taskStoreManager = manager
}

// SetErrorHandler sets the error handler.
func (queue *TasksQueue[D]) SetErrorHandler(handler ErrorHandler) {
	queue.errorHandlerMu.Lock()
	defer queue.errorHandlerMu.Unlock()
	queue.errorHandler = handler
}
