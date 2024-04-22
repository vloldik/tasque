package tasque

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockTaskStorage struct{}

func (m *MockTaskStorage) SaveTaskToStorage(ctx context.Context, data *interface{}) error {
	return nil
}

func (m *MockTaskStorage) GetTasksFromStorage(ctx context.Context, count int) ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *MockTaskStorage) DeleteTaskFromStorage(ctx context.Context, data *interface{}) error {
	return nil
}

func MockErrorHandler(err error) bool {
	return false
}

func MockTask(ctx context.Context, data interface{}) bool {
	return true
}

func TestCreateTaskQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

	if queue.threadCount != 3 {
		t.Error("Expected threadCount to be 3")
	}

	if queue.taskStorageManager != taskStorageManager {
		t.Error("Expected taskStorageManager to be set correctly")
	}

	if queue.task == nil {
		t.Error("Expected task to be set")
	}

	if queue.errorHandler == nil {
		t.Error("Expected errorHandler to be set correctly")
	}

	if queue.signalChannel == nil {
		t.Error("Expected signalChannel to be initialized")
	}
}

func TestSendToQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

	err := queue.SendToQueue(context.Background(), struct{}{})

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}
}

func TestGetFreeThreadsCount(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

	queue.StartQueue(context.Background())
}

func TestGetFromQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

	data := queue.getFromQueue(context.Background())

	if len(data) != 0 {
		t.Errorf("Expected data to be empty, got %v", data)
	}
}

func TestStartQueue(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

	err := queue.StartQueue(context.Background())

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}
}

func TestStartQueue_AlreadyStarted(t *testing.T) {
	taskStorageManager := &MockTaskStorage{}
	errorHandler := MockErrorHandler

	queue := CreateTaskQueue(MockTask, 3, taskStorageManager, errorHandler, time.Hour)

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

func TestQueueUsage(t *testing.T) {
	taskStorageManager := NewMemoryTaskStorage[int]()
	errorHandler := func(err error) bool {
		fmt.Printf("Error %e\n", err)
		return true
	}

	wg := sync.WaitGroup{}
	wg.Add(10)
	ctx := context.Background()

	queue := CreateTaskQueue(func(ctx context.Context, data int) (needToReshedule bool) {
		time.Sleep(time.Second * time.Duration(data))
		if rand.Float64() > 0.5 {
			panic(errors.New("test"))
		}
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Printf("Completed %d\n", data)
			wg.Done()
		}
		return true
	}, 3, taskStorageManager, errorHandler, time.Second*3)

	// Start the queue
	err := queue.StartQueue(context.Background())
	for i := 0; i < 10; i++ {
		err := queue.SendToQueue(ctx, rand.Intn(2)+1)
		assert.Nil(t, err)
	}
	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}

	wg.Wait()
}
