package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/google/uuid"
)

type (
	Executable func() error
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	startTime := time.Now()

	wheel := htw.NewTimingWheel[Executable](1*time.Second, startTime, 60)
	registry := htw.NewTaskRegistry[uuid.UUID, Executable](16)

	out := make(chan Executable)
	defer close(out)

	go func() {
		for executionBlock := range out {
			if err := executionBlock(); err != nil {
				log.Println("Execution error:", err)
			}
		}
	}()

	clock := htw.NewTickerClock(time.Second)
	defer clock.Stop()

	engine := htw.NewEngine(wheel, clock, out)
	go func() {
		if err := engine.Run(ctx); errors.Is(err, context.Canceled) {
			return
		} else if err != nil {
			panic(err)
		}
	}()

	taskID := uuid.New()
	initialExpiry := startTime.Add(3 * time.Second)

	initialTask := htw.NewTask[Executable](initialExpiry, func() error {
		fmt.Println("This shouldn't print if rescheduled successfully!")
		return nil
	})

	if node := wheel.Add(initialTask); node != nil {
		registry.Add(taskID, node)
	}

	time.Sleep(1 * time.Second)

	if oldNode := registry.Remove(taskID); oldNode != nil {
		wheel.Remove(oldNode)
	}

	newExpiry := time.Now().Add(7 * time.Second)
	updatedTask := htw.NewTask[Executable](newExpiry, func() error {
		fmt.Println("Success! The updated task was executed at its new prolonged time.")
		return nil
	})

	if newNode := wheel.Add(updatedTask); newNode != nil {
		registry.Add(taskID, newNode)
	}

	<-ctx.Done()

	log.Println("Engine exited.")
}
