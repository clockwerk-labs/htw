package htw

import (
	"context"
	"sync"
)

type Engine[T any] struct {
	wheel *TimingWheel[T]
	clock Clock
	out   chan T
	mu    sync.Mutex
}

func NewEngine[T any](wheel *TimingWheel[T], clock Clock, out chan T) *Engine[T] {
	return &Engine[T]{
		wheel: wheel,
		clock: clock,
		out:   out,
	}
}

func (e *Engine[T]) Run(ctx context.Context) error {
	tickChan := e.clock.TickChan()
	defer e.clock.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tickChan:
			e.advance()
		}
	}
}

func (e *Engine[T]) advance() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if dueTasks := e.wheel.AdvanceClock(e.clock.Now()); len(dueTasks) > 0 {
		for _, task := range dueTasks {
			e.out <- task.Value
		}
	}
}
