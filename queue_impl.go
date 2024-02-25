package tasque

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrQueueIsFull = errors.New("message queue is full")

type TasksQueue[T Task] struct {
	channel          chan T
	cacheCapacity    int
	errorHandler     ErrorHandler
	taskStoreManager TaskStorageSaveGetDeleter[T]
	errorHandlerMu   sync.Mutex
}

func NewTasksQueue[T Task](cacheCapacity int) *TasksQueue[T] {
	return &TasksQueue[T]{
		channel:       make(chan T, cacheCapacity),
		cacheCapacity: cacheCapacity,
		errorHandler:  func(err error) {},
	}
}

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
				errHandlerFunc = func(err error) {}
			}
		}
	}
}

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

func (queue *TasksQueue[T]) StartQueue(ctx context.Context) {
	go queue.startQueue(ctx)
}

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

func (queue *TasksQueue[T]) SetTaskStoreManager(manager TaskStorageSaveGetDeleter[T]) {
	queue.taskStoreManager = manager
}

func (queue *TasksQueue[T]) SetErrorHandler(handler ErrorHandler) {
	queue.errorHandlerMu.Lock()
	defer queue.errorHandlerMu.Unlock()
	queue.errorHandler = handler
}
