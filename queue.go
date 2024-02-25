package tasque

import "context"

type Task interface {
	Do(ctx context.Context)
}

type ErrorHandler func(err error)

type ErrorHandlerSetter interface {
	SetErrorHandler(handler ErrorHandler)
}

type TaskStorageSaveGetDeleter[T Task] interface {
	SaveTaskToStorage(ctx context.Context, task *T) error

	// Returns an array with lenght [TaskFromStorageBatchCount()]
	GetTasksFromStorage(ctx context.Context) ([]T, error)

	DeleteTaskFromStorage(ctx context.Context, task *T) error

	// Returns lenght of single [GetAndDeleteTasksFromStorage]
	TaskFromStorageBatchCount() int
}

type TasksQueueManger[T Task] interface {
	StartQueue(ctx context.Context)
	SendToQueue(ctx context.Context, item T) error
}

type TasksQueueManagerWithStoreAndErrorHandler[T Task] interface {
	TasksQueueManger[T]
	SetTaskStoreManager(manager TaskStorageSaveGetDeleter[T])
	ErrorHandlerSetter
}
