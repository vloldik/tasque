package tasque

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
)

type Tasque[D struct{}] struct {
	threadCount        int
	task               Task[D]
	taskStorageManager TaskStorageSaveGetDeleter[D]
	currentlyRunning   atomic.Int32
	signalChannel      chan struct{}
	errorHandler       ErrorHandler
	isStarted          atomic.Bool
}

func CreateTaskQueue[D struct{}](task Task[D], threadCount int, taskStorageManager TaskStorageSaveGetDeleter[D],
	errorHandler ErrorHandler) *Tasque[D] {
	return &Tasque[D]{
		threadCount:        threadCount,
		taskStorageManager: taskStorageManager,
		task:               task,
		errorHandler:       errorHandler,
		signalChannel:      make(chan struct{}),
	}
}

func (queue *Tasque[D]) SendToQueue(ctx context.Context, data D) (err error) {
	err = queue.taskStorageManager.SaveTaskToStorage(ctx, &data)
	go func() { queue.signalChannel <- struct{}{} }()
	return
}

func (queue *Tasque[D]) getFreeThreadsCount() int {
	return queue.threadCount - int(queue.currentlyRunning.Load())
}

func (queue *Tasque[D]) getFromQueue(ctx context.Context) (data []D) {
	data, err := queue.taskStorageManager.GetTasksFromStorage(ctx, queue.getFreeThreadsCount())
	if err != nil {
		log.Fatalf("An error occured while reading tasks from memory! %e", err)
	}
	return data
}

func (queue *Tasque[D]) startTask(ctx context.Context, data D) {
	queue.currentlyRunning.Add(1)
	needToReshedule := false
	defer func() {
		queue.currentlyRunning.Add(-1)
		queue.signalChannel <- struct{}{}
		if r := recover(); r != nil {
			needToReshedule = queue.errorHandler(fmt.Errorf("panic occurred: %v", r))
		}
		queue.taskStorageManager.DeleteTaskFromStorage(ctx, &data)
		if needToReshedule {
			queue.SendToQueue(ctx, data)
		}
	}()
	needToReshedule = queue.task(ctx, data)
}

func (queue *Tasque[D]) startQueue(ctx context.Context) {
	for {
		select {
		case <-queue.signalChannel:
			if queue.getFreeThreadsCount() == 0 {
				break
			}
			tasks := queue.getFromQueue(ctx)
			for _, data := range tasks {
				go queue.startTask(ctx, data)
			}
		case <-ctx.Done():
			return

		}
	}
}

func (queue *Tasque[D]) StartQueue(ctx context.Context) error {
	if queue.isStarted.Load() {
		return errors.New("queue is already started")
	}
	queue.isStarted.Store(true)
	go queue.startQueue(ctx)
	return nil
}
