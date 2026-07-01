package htw_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/stretchr/testify/require"
)

type FakeClock struct {
	currentTime time.Time
	tickChan    chan time.Time
	stopped     bool
}

func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{
		currentTime: start,
		tickChan:    make(chan time.Time, 1),
	}
}

func (f *FakeClock) Now() time.Time {
	return f.currentTime
}

func (f *FakeClock) TickChan() <-chan time.Time {
	return f.tickChan
}

func (f *FakeClock) Stop() {
	f.stopped = true
}

func (f *FakeClock) Tick(d time.Duration) {
	f.currentTime = f.currentTime.Add(d)

	if !f.stopped {
		f.tickChan <- f.currentTime
	}
}

func TestEngine_DeterministicExecution(t *testing.T) {
	startTime := time.Unix(1000, 0)
	fakeClock := NewFakeClock(startTime)

	tickDuration := 10 * time.Millisecond
	wheel := htw.NewTimingWheel[string](tickDuration, startTime, 10)
	outChan := make(chan string, 5)

	engine := htw.NewEngine(wheel, fakeClock, outChan)

	task1 := htw.NewTask(startTime.Add(20*time.Millisecond), "task-1")
	task2 := htw.NewTask(startTime.Add(40*time.Millisecond), "task-2")

	require.NotNil(t, wheel.Add(task1))
	require.NotNil(t, wheel.Add(task2))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engineErrChan := make(chan error, 1)
	go func() {
		engineErrChan <- engine.Run(ctx)
	}()

	fakeClock.Tick(tickDuration)
	select {
	case val := <-outChan:
		t.Fatalf("Unexpected task execution early: %s", val)
	case <-time.After(5 * time.Millisecond):
	}

	fakeClock.Tick(tickDuration)
	select {
	case val := <-outChan:
		require.Equal(t, "task-1", val)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timeout waiting for task-1")
	}

	fakeClock.Tick(tickDuration)
	fakeClock.Tick(tickDuration)
	select {
	case val := <-outChan:
		require.Equal(t, "task-2", val)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timeout waiting for task-2")
	}

	cancel()
	select {
	case err := <-engineErrChan:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Engine failed to exit when context was canceled")
	}
}

func BenchmarkEngine_Throughput(b *testing.B) {
	startTime := time.Now()
	wheel := htw.NewTimingWheel[int](1*time.Millisecond, startTime, 100)
	outChan := make(chan int, b.N)

	for i := range b.N {
		wheel.Add(htw.NewTask(startTime.Add(5*time.Millisecond), i))
	}

	mockClock := NewFakeClock(startTime)
	engine := htw.NewEngine(wheel, mockClock, outChan)

	go func() {
		if err := engine.Run(b.Context()); errors.Is(err, context.Canceled) {
			return
		} else if err != nil {
			b.Error(err)
		}
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		count := 0
		for range outChan {
			count++

			if count == b.N {
				return
			}
		}
	})

	for range 10 {
		mockClock.Tick(1 * time.Millisecond)
	}

	wg.Wait()
}
