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

	idToRemove := uuid.New()

	if node := wheel.Add(htw.NewTask[Executable](startTime.Add(5*time.Second), func() error {
		fmt.Printf("Hello World %s!\n", "removed")

		return nil
	})); node != nil {
		registry.Add(idToRemove, node)
	}

	for i := 0; i < 4; i++ {
		if node := wheel.Add(htw.NewTask[Executable](startTime.Add(5*time.Second), func() error {
			fmt.Printf("Hello World %d!\n", i)

			return nil
		})); node != nil {
			registry.Add(uuid.New(), node)
		}
	}

	out := make(chan Executable)
	defer close(out)

	go func() {
		for o := range out {
			if err := o(); err != nil {
				log.Println(err)
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

	if nodeToRemove, ok := registry.Get(idToRemove); ok {
		wheel.Remove(nodeToRemove)
		registry.Remove(idToRemove)
	}

	<-ctx.Done()

	log.Println("Engine exited")
}
