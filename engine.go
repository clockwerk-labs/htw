package htw

import (
	"context"
	"sync"
	"time"
)

type Engine[T any] struct {
	wheel *TimingWheel[T]
	out   chan T
	mu    sync.Mutex
}

func NewEngine[T any](wheel *TimingWheel[T], out chan T) *Engine[T] {
	return &Engine[T]{
		wheel: wheel,
		out:   out,
	}
}

func (e *Engine[T]) Run(ctx context.Context, tick time.Duration) error {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			e.advance()
		}
	}
}

func (e *Engine[T]) advance() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if dueTasks := e.wheel.AdvanceClock(time.Now()); len(dueTasks) > 0 {
		for _, task := range dueTasks {
			e.out <- task.Value
		}
	}

}
