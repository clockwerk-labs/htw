package htw

import (
	"hash/maphash"
	"iter"
	"sync"
)

type (
	Registry[K comparable, T any] struct {
		mask   uint64
		shards []*registryShard[K, T]
		seed   maphash.Seed
	}

	registryShard[K comparable, T any] struct {
		items map[K]T
		mu    sync.RWMutex
	}
)

func NewRegistry[K comparable, T any](shardCount int) *Registry[K, T] {
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 32
	}

	shards := make([]*registryShard[K, T], shardCount)
	for i := range shardCount {
		shards[i] = &registryShard[K, T]{
			items: make(map[K]T),
		}
	}

	return &Registry[K, T]{
		mask:   uint64(shardCount - 1),
		shards: shards,
		seed:   maphash.MakeSeed(),
	}
}

func (tr *Registry[K, T]) Add(key K, node T) {
	tr.getShard(key).Add(key, node)
}

func (tr *Registry[K, T]) Remove(key K) (T, bool) {
	return tr.getShard(key).Remove(key)
}

func (tr *Registry[K, T]) Get(key K) (T, bool) {
	return tr.getShard(key).Get(key)
}

func (tr *Registry[K, T]) Clear() {
	for _, shard := range tr.shards {
		shard.Clear()
	}
}

func (tr *Registry[K, T]) GetAll() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, shard := range tr.shards {
			if !shard.StreamShard(yield) {
				return
			}
		}
	}
}

func (tr *Registry[K, T]) getShard(key K) *registryShard[K, T] {
	return tr.shards[maphash.Comparable(tr.seed, key)&tr.mask]
}

func (s *registryShard[K, T]) Add(key K, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = value
}

func (s *registryShard[K, T]) Remove(key K) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.items[key]
	if !ok {
		var zero T
		return zero, false
	}

	delete(s.items, key)

	return node, true
}

func (s *registryShard[K, T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[K]T)
}

func (s *registryShard[K, T]) Get(key K) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.items[key]

	return t, ok
}

func (s *registryShard[K, T]) StreamShard(yield func(T) bool) bool {
	snapshot := s.Snapshot()

	for _, t := range snapshot {
		if !yield(t) {
			return false
		}
	}

	return true
}

func (s *registryShard[K, T]) Snapshot() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make([]T, 0, len(s.items))
	for _, t := range s.items {
		snapshot = append(snapshot, t)
	}

	return snapshot
}
