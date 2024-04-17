package tasque

import (
	"context"
	"sync"
)

type MemoryTaskStorage[D struct{}] struct {
	mutex     sync.Mutex
	taskQueue []D
}

func NewMemoryTaskStorage[D struct{}]() *MemoryTaskStorage[D] {
	return &MemoryTaskStorage[D]{
		taskQueue: make([]D, 0),
	}
}

func (storage *MemoryTaskStorage[D]) SaveTaskToStorage(ctx context.Context, data *D) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	storage.taskQueue = append(storage.taskQueue, *data)
	return nil
}

func (storage *MemoryTaskStorage[D]) GetTasksFromStorage(ctx context.Context, count int) ([]D, error) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if count > len(storage.taskQueue) {
		count = len(storage.taskQueue)
	}

	data := make([]D, count)
	copy(data, storage.taskQueue[:count])

	return data, nil
}

func (storage *MemoryTaskStorage[D]) DeleteTaskFromStorage(ctx context.Context, data *D) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	for i, task := range storage.taskQueue {
		if task == *data {
			storage.taskQueue = append(storage.taskQueue[:i], storage.taskQueue[i+1:]...)
			break
		}
	}

	return nil
}
