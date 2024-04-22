package tasque

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrTimeout = errors.New("timeout error")

type SaveTaskError error
type DeleteTaskError error
type ReadTaskError error

func CreateReadTaskError(orig error) SaveTaskError {
	return SaveTaskError(orig)
}

func CreateDeleteTaskError(orig error) DeleteTaskError {
	return DeleteTaskError(orig)
}

func CreateSaveTaskError(orig error) SaveTaskError {
	return SaveTaskError(orig)
}

type Tasque[D interface{}] struct {
	threadCount        int
	task               Task[D]
	taskStorageManager TaskStorageSaveGetDeleter[D]
	signalChannel      chan struct{}
	errorHandler       ErrorHandler
	isStarted          atomic.Bool
	wg                 sync.WaitGroup
	taskTimeout        time.Duration
}

func CreateTaskQueue[D interface{}](task Task[D], threadCount int, taskStorageManager TaskStorageSaveGetDeleter[D],
	errorHandler ErrorHandler, taskTimeout time.Duration) *Tasque[D] {
	return &Tasque[D]{
		threadCount:        threadCount,
		taskStorageManager: taskStorageManager,
		task:               task,
		errorHandler:       errorHandler,
		signalChannel:      make(chan struct{}),
		wg:                 sync.WaitGroup{},
		taskTimeout:        taskTimeout,
	}
}

func (queue *Tasque[D]) sendToQueue(ctx context.Context, notify bool, datas ...D) (err error) {
	for _, data := range datas {
		err = queue.taskStorageManager.SaveTaskToStorage(ctx, &data)
		if err != nil {
			return
		}
	}
	if notify {
		go func() { queue.signalChannel <- struct{}{} }()
	}
	return
}

func (queue *Tasque[D]) SendToQueue(ctx context.Context, datas ...D) (err error) {
	return queue.sendToQueue(ctx, true, datas...)
}

func (queue *Tasque[D]) getFromQueue(ctx context.Context) (data []D) {
	data, err := queue.taskStorageManager.GetTasksFromStorage(ctx, queue.threadCount)
	if err != nil {
		queue.errorHandler(CreateReadTaskError(err))
		return []D{}
	}
	return data
}

func (queue *Tasque[D]) startTask(ctx context.Context, data D) {
	needToResheduleChan := make(chan bool)
	needToReshedule := false
	defer func() {
		err := queue.taskStorageManager.DeleteTaskFromStorage(ctx, &data)
		if err != nil {
			queue.errorHandler(CreateDeleteTaskError(err))
			return
		}
		if !needToReshedule {
			return
		}
		if err := queue.sendToQueue(ctx, false, data); err != nil {
			queue.errorHandler(CreateSaveTaskError(err))
		}
		queue.wg.Done()
	}()
	go func(task Task[D]) {
		defer func() {
			if err := recover(); err != nil {
				needToResheduleChan <- queue.errorHandler(fmt.Errorf("panic occurred: %v", err))
			}
		}()
		needToResheduleChan <- task(ctx, data)
	}(queue.task)
	select {
	case value := <-needToResheduleChan:
		needToReshedule = value
	case <-ctx.Done():
		needToReshedule = queue.errorHandler(ErrTimeout)
	}
}

func (queue *Tasque[D]) startTasks(ctx context.Context) {
	tasks := queue.getFromQueue(ctx)
	if len(tasks) == 0 {
		return
	}
	timeoutCtx, cancelContext := context.WithTimeout(ctx, queue.taskTimeout)
	for _, data := range tasks {
		queue.wg.Add(1)
		go queue.startTask(timeoutCtx, data)
	}
	queue.wg.Wait()
	cancelContext()
	go func() { queue.signalChannel <- struct{}{} }()
}

func (queue *Tasque[D]) startQueue(ctx context.Context) {
	for {
		select {
		case <-queue.signalChannel:
			queue.startTasks(ctx)
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
