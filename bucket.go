package htw

import (
	"sync"
)

type Bucket[T any] struct {
	tasks []Task[T]
	mu    sync.Mutex
}

func NewBucket[T any]() Bucket[T] {
	return Bucket[T]{
		tasks: make([]Task[T], 0),
	}
}

func (b *Bucket[T]) Add(task Task[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tasks = append(b.tasks, task)
}

func (b *Bucket[T]) Flush() []Task[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.tasks) == 0 {
		return nil
	}

	evicted := b.tasks
	b.tasks = make([]Task[T], 0)

	return evicted
}
