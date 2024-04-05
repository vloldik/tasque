package tasque

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define a mock TaskStorageSaveGetDeleter interface for testing purposes.
type MockTaskStorageSaveGetDeleter struct {
	mock.Mock
}

func (m *MockTaskStorageSaveGetDeleter) TaskFromStorageBatchCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTaskStorageSaveGetDeleter) GetTasksFromStorage(ctx context.Context) ([]int, error) {
	args := m.Called(ctx)
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockTaskStorageSaveGetDeleter) SaveTaskToStorage(ctx context.Context, item *int) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockTaskStorageSaveGetDeleter) DeleteTaskFromStorage(ctx context.Context, item *int) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

// Test the NewTasksQueue function.
func TestNewTasksQueue(t *testing.T) {
	queue := NewTasksQueue[int](func(ctx context.Context, data int) {}, 10)
	assert.NotNil(t, queue)
	assert.Equal(t, 10, queue.cacheCapacity)
}

// Test the addBatchFromStorageIfNeeded function.
func TestAddBatchFromStorageIfNeeded(t *testing.T) {
	ctx := context.Background()

	// Create a mocked task storage manager.
	taskStorageManager := new(MockTaskStorageSaveGetDeleter)
	taskStorageManager.On("DeleteTaskFromStorage", ctx, mock.Anything).Return(nil).Once()
	taskStorageManager.On("TaskFromStorageBatchCount").Return(5)
	taskStorageManager.On("GetTasksFromStorage", ctx).
		Return([]int{1}, nil)

	// Create a task queue and set the mocked task storage manager.
	queue := NewTasksQueue[int](Task[int](func(ctx context.Context, data int) {}), 10)
	queue.SetTaskStoreManager(taskStorageManager)

	// Call the addBatchFromStorageIfNeeded function.
	queue.addBatchFromStorageIfNeeded(ctx)

	// Verify that the mocked task storage manager was called.
	taskStorageManager.AssertExpectations(t)
}

// Test the doTarget function.
func TestDoTarget(t *testing.T) {
	ctx := context.Background()
	done := false

	// Create a mocked task.
	task := Task[int](func(ctx context.Context, data int) {
		done = true
	})
	// Create a task queue.
	queue := NewTasksQueue[int](task, 10)

	// Call the doTarget function.
	queue.doTarget(ctx, 42)

	// Verify that the mocked task was called.
	assert.True(t, done)
}

// Test the startQueue function.
func TestStartQueue(t *testing.T) {
	ctx := context.Background()

	done := make(chan bool)

	// Create a mocked task.
	task := func(ctx context.Context, data int) {
		done <- true
	}

	// Create a task queue.
	queue := NewTasksQueue[int](Task[int](task), 10)
	queue.SendToQueue(ctx, 42)

	// Start the task queue in a separate goroutine.
	go func() {
		queue.StartQueue(ctx)
		close(done)
	}()

	// Wait for the task queue to finish.
	<-done

}

func TestMainUsage(t *testing.T) {
	ctx := context.Background()
	storage := MockTaskStorageSaveGetDeleter{}

	countDown := sync.WaitGroup{}
	countDown.Add(4)

	queue := NewTasksQueue(func(ctx context.Context, data int) {
		fmt.Printf("Hello from queue! Item №%d\n", data)
		time.Sleep(time.Second)
		countDown.Done()
	}, 2)
	storage.On("DeleteTaskFromStorage", ctx, mock.Anything).Return(nil)
	storage.On("SaveTaskToStorage", ctx, mock.Anything).Return(nil)
	storage.On("TaskFromStorageBatchCount").Return(1)
	storage.On("GetTasksFromStorage", ctx).
		Return([]int{3, 4}, nil)

	queue.SetTaskStoreManager(&storage)
	queue.SetErrorHandler(func(err error) { fmt.Printf("An error occurred %e\n", err) })
	queue.SendToQueue(ctx, 1)
	queue.SendToQueue(ctx, 2)
	queue.StartQueue(ctx)
	countDown.Wait()
}
