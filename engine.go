package htw

import (
	"context"
)

type Engine[T any] struct {
	wheel *TimingWheel[T]
	clock Clock
	out   chan T
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
	if dueNodes := e.wheel.AdvanceClock(e.clock.Now()); len(dueNodes) > 0 {
		for _, node := range dueNodes {
			e.out <- node.Task.Value
		}
	}
}
