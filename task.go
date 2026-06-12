package ratukas

import (
	"fmt"
	"sync"
	"time"
)

type (
	Task[T any] struct {
		expiration int64
		value      T
	}

	TaskRegistry[T any] struct {
		shards []*taskShard[T]
	}

	taskShard[T any] struct {
		tasks map[uint64]*Task[T]
		mu    sync.RWMutex
	}
)

func NewTask[T any](expiration time.Time, value T) *Task[T] {
	return &Task[T]{
		expiration: expiration.UnixMilli(),
		value:      value,
	}
}

func NewTaskRegistry[T any](noOfShards int) *TaskRegistry[T] {
	shards := make([]*taskShard[T], noOfShards)

	for i := 0; i < noOfShards; i++ {
		shards[i] = &taskShard[T]{
			tasks: make(map[uint64]*Task[T]),
		}
	}

	return &TaskRegistry[T]{
		shards: shards,
	}
}

func (r *TaskRegistry[T]) Get(key uint64) (*Task[T], error) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, ok := s.tasks[key]; !ok {
		return nil, fmt.Errorf("task %d not found", key)
	} else {
		return task, nil
	}
}

func (r *TaskRegistry[T]) Put(key uint64, task *Task[T]) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[key] = task
}

func (r *TaskRegistry[T]) Delete(key uint64) {
	s := r.shards[key%uint64(len(r.shards))]

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, key)
}
