package htw

import (
	"time"
)

type (
	Task[T any] struct {
		Expiration int64
		Value      T
	}
)

func NewTask[T any](expiration time.Time, value T) Task[T] {
	return Task[T]{
		Expiration: expiration.UnixNano(),
		Value:      value,
	}
}
