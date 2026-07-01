package main

import (
	"context"
	"errors"
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

	wheel := htw.NewTimingWheel[Executable](1*time.Second, time.Now(), 60)
	registry := htw.NewRegistry[uuid.UUID, *htw.Node[Executable]](16)

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

	initialTask := htw.NewTask[Executable](time.Now().Add(3*time.Second), func() error {
		log.Println("This shouldn't print if rescheduled successfully!")

		return nil
	})

	if node := wheel.Add(initialTask); node != nil {
		registry.Add(taskID, node)

		log.Println("Scheduled task", taskID.String())
	}

	time.Sleep(1 * time.Second)

	if oldNode, ok := registry.Remove(taskID); ok {
		if wheel.Remove(oldNode) {
			log.Printf("Removed task %s", taskID.String())
		}
	}

	updatedTask := htw.NewTask[Executable](time.Now().Add(7*time.Second), func() error {
		log.Println("Success! The updated task was executed at its new prolonged time.")

		return nil
	})

	if newNode := wheel.Add(updatedTask); newNode != nil {
		registry.Add(taskID, newNode)

		log.Println("Rescheduled task", taskID.String())
	}

	<-ctx.Done()

	log.Println("Engine exited.")
}
