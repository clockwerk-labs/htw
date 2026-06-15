# Hierarchical Timing Wheels (HTW)

[![Go Version](https://img.shields.io/badge/go-1.26.4-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A highly efficient, zero-dependency, thread-safe Hierarchical Timing Wheel implementation in Go. This library is designed for scheduling and tracking millions of timers or low-latency asynchronous tasks with minimal CPU and memory overhead compared to standard language mechanisms like `time.After` or `time.NewTimer`.

---

## Why Use Hierarchical Timing Wheels?

Standard timers in many runtimes (including Go's runtime timer heap) rely on a priority queue structure, scaling with a time complexity of $O(\log N)$ for insertion and deletion. While highly accurate, this can bottleneck performance when handling millions of concurrent connections, open sockets, or transient events.

This implementation reduces those costs to **$O(1)$ amortized time** for insertions, deletions, and executions by utilizing a fixed-size ring buffer of buckets. When a task falls beyond the time horizon of the primary wheel, it seamlessly overflows into a higher-level coarser wheel, cascading downward automatically as time ticks forward.

---

## Installation

Ensure you have Go installed (version `1.26.4` or higher is specified for this workspace).

```bash
go get [github.com/clockwerk-labs/htw](https://github.com/clockwerk-labs/htw)
```

## Quick Start

Below is a complete example demonstrating how to initialize the engine, spin up a processing worker, schedule tasks, and handle termination gracefully.

```go
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

type Executable func() error

func main() {
	// 1. Create a cancellable context matching system termination signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	startTime := time.Now()

	// 2. Instantiate a Timing Wheel: 1-second ticks, 60 slots (1-minute coverage range per level)
	wheel := htw.NewTimingWheel[Executable](1*time.Second, startTime, 60)

	// 3. Create your outbound execution channel
	out := make(chan Executable)
	defer close(out)

	// 4. Spin up workers to consume due tasks
	go func() {
		for executionBlock := range out {
			if err := executionBlock(); err != nil {
				log.Println("Task execution error:", err)
			}
		}
	}()

	// 5. Connect a system clock source and run the engine
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

	// 6. Schedule some future items
	for i := 0; i < 100; i++ {
		taskIndex := i
		task := htw.NewTask[Executable](uuid.New(), startTime.Add(5*time.Second), func() error {
			fmt.Printf("Executed Task #%d!\n", taskIndex)
			return nil
		})

		if ok := wheel.Add(task); !ok {
			log.Println("Task dropped: expiration window already passed.")
		}
	}

	// Wait for Ctrl+C / Termination
	<-ctx.Done()
	log.Println("Engine exited gracefully.")
```

## Architecture & Component Design

The project architecture is composed of distinct layers separating data structural calculations, concurrency synchronization, and system time driving loops.
Component Overview

 - Task[T]: The core wrapper package holding a generic payload, explicit expiration timestamp (converted internally to nanoseconds), and a distinct uuid.UUID identifier.
 - Bucket[T]: A thread-safe, mutex-protected array layout grouping tasks destined for the exact same granular tick window.
 - TimingWheel[T]: The circular array structural component representing time. Manages routing logic, tracking boundaries, and dynamically mounts pointer allocations to its overflow wheel properties.
 - Engine[T]: The orchestration wrapper bridging an abstracted Clock interface ticker loop to the underlying state machine.

## The Overflow & Cascade Mechanism

```
[ Primary Wheel ] (e.g., 10ms Ticks, 10 Slots -> 100ms range)
   ↳ Slot 0, 1, 2... 9
   ↳ [Task Expiring @ +40ms]  -> Placed in Slot 4.
   ↳ [Task Expiring @ +250ms] -> Exceeds range! Triggers Overflow creation.
        │
        ▼
[ Overflow Wheel Level 1 ] (100ms Ticks, 10 Slots -> 1000ms range)
   ↳ Slot 0, 1, 2... 9
   ↳ Placed in Slot 2 (+200ms to +300ms window).
```

When AdvanceClock is invoked by the Engine loop, the primary wheel drains its active bucket. It then recursively commands its overflow wheels to advance. Tasks residing in higher levels whose expiration timestamps now fit back within the newly advanced primary boundary are automatically flushed and re-routed downward into the tighter primary wheels.

## Testing

The codebase includes strong automated test coverage asserting data integrity, determinism via mock clocks, and thread safety.

To run the complete test suite, execute:

```bash
go test -v ./...
```
