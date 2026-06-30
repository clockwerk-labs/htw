package htw

import (
	"hash/maphash"
	"iter"
	"sync"
	"time"
)

type (
	Task[T any] struct {
		Expiration int64
		Value      T
	}

	TaskRegistry[K comparable, T any] struct {
		mask   uint32
		shards []*registryShard[K, T]
		seed   maphash.Seed
	}

	registryShard[K comparable, T any] struct {
		items map[K]*Node[T]
		mu    sync.RWMutex
	}
)

func NewTask[T any](expiration time.Time, value T) Task[T] {
	return Task[T]{
		Expiration: expiration.UnixNano(),
		Value:      value,
	}
}

func NewTaskRegistry[K comparable, T any](shardCount int) *TaskRegistry[K, T] {
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 32
	}

	shards := make([]*registryShard[K, T], shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &registryShard[K, T]{
			items: make(map[K]*Node[T]),
		}
	}

	return &TaskRegistry[K, T]{
		mask:   uint32(shardCount - 1),
		shards: shards,
		seed:   maphash.MakeSeed(),
	}
}

func (tr *TaskRegistry[K, T]) Add(key K, node *Node[T]) {
	tr.getShard(key).Add(key, node)
}

func (tr *TaskRegistry[K, T]) Remove(key K) *Node[T] {
	return tr.getShard(key).Remove(key)
}

func (tr *TaskRegistry[K, T]) Get(key K) (*Node[T], bool) {
	return tr.getShard(key).Get(key)
}

func (tr *TaskRegistry[K, T]) Clear() {
	for _, shard := range tr.shards {
		shard.Clear()
	}
}

func (tr *TaskRegistry[K, T]) GetAll() iter.Seq[*Node[T]] {
	return func(yield func(*Node[T]) bool) {
		for _, shard := range tr.shards {
			if !shard.StreamShard(yield) {
				return
			}
		}
	}
}

func (tr *TaskRegistry[K, T]) getShard(key K) *registryShard[K, T] {
	return tr.shards[uint32(maphash.Comparable(tr.seed, key))&tr.mask]
}

func (s *registryShard[K, T]) Add(key K, node *Node[T]) {
	if node == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = node
}

func (s *registryShard[K, T]) Remove(key K) *Node[T] {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.items[key]
	if !ok {
		return nil
	}

	delete(s.items, key)

	return node
}

func (s *registryShard[K, T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[K]*Node[T])
}

func (s *registryShard[K, T]) Get(key K) (*Node[T], bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.items[key]

	return t, ok
}

func (s *registryShard[K, T]) StreamShard(yield func(*Node[T]) bool) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.items {
		if !yield(t) {
			return false
		}
	}

	return true
}
