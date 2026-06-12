package htw

import (
	"time"
)

type TimingWheel[T any] struct {
	tick        int64
	size        int64
	interval    int64
	currentTime int64
	buckets     []Bucket[T]
	overflow    *TimingWheel[T]
}

func NewTimingWheel[T any](tick time.Duration, start time.Time, size int64) *TimingWheel[T] {
	tickNs, startNs := tick.Nanoseconds(), start.UnixNano()

	buckets := make([]Bucket[T], size)
	for i := range buckets {
		buckets[i] = NewBucket[T]()
	}

	return &TimingWheel[T]{
		tick:        tickNs,
		size:        size,
		interval:    tickNs * size,
		currentTime: startNs - (startNs % tickNs),
		buckets:     buckets,
	}
}

func (tw *TimingWheel[T]) Add(task Task[T]) bool {
	if task.Expiration < tw.currentTime+tw.tick {
		return false
	}

	if task.Expiration < tw.currentTime+tw.interval {
		tw.buckets[(task.Expiration/tw.tick)%tw.size].Add(task)

		return true
	}

	if tw.overflow == nil {
		tw.overflow = NewTimingWheel[T](time.Duration(tw.interval), time.Unix(0, tw.currentTime), tw.size)
	}

	return tw.overflow.Add(task)
}

func (tw *TimingWheel[T]) AdvanceClock(targetTime time.Time) (dueTasks []Task[T]) {
	targetTimeNs := targetTime.UnixNano()

	for targetTimeNs >= tw.currentTime+tw.tick {
		tw.currentTime += tw.tick

		dueTasks = append(dueTasks, tw.buckets[(tw.currentTime/tw.tick)%tw.size].Flush()...)

		if tw.overflow != nil {
			overflowTasks := tw.overflow.AdvanceClock(time.Unix(0, tw.currentTime))

			for _, task := range overflowTasks {
				if !tw.Add(task) {
					dueTasks = append(dueTasks, task)
				}
			}
		}
	}

	return dueTasks
}
