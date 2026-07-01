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

	registry := htw.NewRegistry[uuid.UUID, *htw.Node[Executable]](16)

	idToRemove := uuid.New()

	if node := wheel.Add(htw.NewTask[Executable](startTime.Add(5*time.Second), func() error {
		fmt.Printf("Hello World %s!", "removed")

		return nil
	})); node != nil {
		registry.Add(idToRemove, node)

		log.Printf("Added %s", idToRemove)
	}

	for i := range 4 {
		if node := wheel.Add(htw.NewTask[Executable](startTime.Add(5*time.Second), func() error {
			log.Printf("Hello World %d!", i)

			return nil
		})); node != nil {
			registry.Add(uuid.New(), node)

			log.Printf("Added %s", idToRemove)
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

	if nodeToRemove, ok := registry.Remove(idToRemove); ok {
		if wheel.Remove(nodeToRemove) {
			log.Printf("Removed %s", idToRemove.String())
		}
	}

	<-ctx.Done()

	log.Println("Engine exited")
}
