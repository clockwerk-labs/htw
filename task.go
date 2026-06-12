package htw

import (
	"time"

	"github.com/google/uuid"
)

type Task[T any] struct {
	Id         uuid.UUID
	Expiration int64
	Value      T
}

func NewTask[T any](id uuid.UUID, expiration time.Time, value T) Task[T] {
	return Task[T]{
		Id:         id,
		Expiration: expiration.UnixNano(),
		Value:      value,
	}
}
