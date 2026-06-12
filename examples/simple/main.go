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

	out := make(chan Executable)
	defer close(out)

	go func() {
		for o := range out {
			if err := o(); err != nil {
				log.Println(err)
			}
		}
	}()

	engine := htw.NewEngine(wheel, out)

	go func() {
		if err := engine.Run(ctx, time.Second); errors.Is(err, context.Canceled) {
			return
		} else if err != nil {
			panic(err)
		}
	}()

	task := htw.NewTask[Executable](uuid.New(), startTime.Add(5*time.Second), func() error {
		fmt.Println("Hello World")

		return nil
	})

	if ok := wheel.Add(task); ok {
		log.Println("Scheduled task")
	} else {
		log.Println("Task not scheduled")
	}

	<-ctx.Done()
}
