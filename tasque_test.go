package tasque

import (
	"context"
	"testing"
)

type MockTaskStorage struct{}

func (m *MockTaskStorage) SaveTaskToStorage(ctx context.Context, data *struct{}) error {
	return nil
}

func (m *MockTaskStorage) GetTasksFromStorage(ctx context.Context, count int) ([]struct{}, error) {
	return []struct{}{}, nil
}

func (m *MockTaskStorage) DeleteTaskFromStorage(ctx context.Context, data *struct{}) error {
	return nil
}

func MockErrorHandler(err error) bool {
	return false
}

func MockTask(ctx context.Context, data struct{}) bool {
	return true
}

func TestCreateTaskQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	if queue.threadCount != 3 {
		t.Error("Expected threadCount to be 3")
	}

	if queue.taskStorageManager != taskStorageManager {
		t.Error("Expected taskStorageManager to be set correctly")
	}

	if queue.task == nil {
		t.Error("Expected task to be set")
	}

	if queue.errorHandler != nil {
		t.Error("Expected errorHandler to be set correctly")
	}

	if queue.signalChannel == nil {
		t.Error("Expected signalChannel to be initialized")
	}
}

func TestSendToQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	err := queue.SendToQueue(context.Background(), struct{}{})

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}
}

func TestGetFreeThreadsCount(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	// Assuming no tasks are currently running
	count := queue.getFreeThreadsCount()

	if count != 3 {
		t.Errorf("Expected count to be 3, got %d", count)
	}
}

func TestGetFromQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	data := queue.getFromQueue(context.Background())

	if len(data) != 0 {
		t.Errorf("Expected data to be empty, got %v", data)
	}
}

func TestStartQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	err := queue.StartQueue(context.Background())

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}
}

func TestStartQueue_AlreadyStarted(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler)

	// Start the queue
	err := queue.StartQueue(context.Background())
	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}

	// Try starting the queue again
	err = queue.StartQueue(context.Background())
	if err == nil || err.Error() != "queue is already started" {
		t.Errorf("Expected err to be 'queue is already started', got %v", err)
	}
}
