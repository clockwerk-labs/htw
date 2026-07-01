package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/clockwerk-labs/htw"
)

type (
	Executable func() error
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	startTime := time.Now()
	wheel := htw.NewTimingWheel[Executable](1*time.Second, startTime, 60)

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

	for i := range 5 {
		task := htw.NewTask[Executable](startTime.Add(5*time.Second), func() error {
			log.Printf("Hello World %d!", i)

			return nil
		})

		if node := wheel.Add(task); node != nil {
			log.Printf("Scheduled task %d", i)
		} else {
			log.Printf("Task %d not scheduled", i)
		}
	}

	<-ctx.Done()

	log.Println("Engine exited")
}
