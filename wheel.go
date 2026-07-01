package htw

import (
	"sync"
	"time"
)

type TimingWheel[T any] struct {
	tick        int64
	size        int64
	interval    int64
	currentTime int64
	buckets     []*Bucket[T]
	overflow    *TimingWheel[T]
	mu          sync.Mutex
}

func NewTimingWheel[T any](tick time.Duration, start time.Time, size int64) *TimingWheel[T] {
	tickNs, startNs := tick.Nanoseconds(), start.UnixNano()

	buckets := make([]*Bucket[T], size)
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

func (tw *TimingWheel[T]) Add(task Task[T]) *Node[T] {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.add(task)
}

func (tw *TimingWheel[T]) Remove(node *Node[T]) bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.remove(node)
}

func (tw *TimingWheel[T]) AdvanceClock(targetTime time.Time) []*Node[T] {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.advanceClock(targetTime)
}

func (tw *TimingWheel[T]) add(task Task[T]) *Node[T] {
	if task.Expiration < tw.currentTime+tw.tick {
		return nil
	}

	if task.Expiration < tw.currentTime+tw.interval {
		return tw.buckets[(task.Expiration/tw.tick)%tw.size].Add(task)
	}

	if tw.overflow == nil {
		tw.overflow = NewTimingWheel[T](time.Duration(tw.interval), time.Unix(0, tw.currentTime), tw.size)
	}

	return tw.overflow.add(task)
}

func (tw *TimingWheel[T]) remove(node *Node[T]) bool {
	if node == nil {
		return false
	}

	if node.Task.Expiration < tw.currentTime+tw.interval {
		return tw.buckets[(node.Task.Expiration/tw.tick)%tw.size].Remove(node)
	}

	if tw.overflow != nil {
		return tw.overflow.remove(node)
	}

	return false
}

func (tw *TimingWheel[T]) advanceClock(targetTime time.Time) (dueNodes []*Node[T]) {
	targetTimeNs := targetTime.UnixNano()

	for targetTimeNs >= tw.currentTime+tw.tick {
		tw.currentTime += tw.tick

		dueNodes = append(dueNodes, tw.buckets[(tw.currentTime/tw.tick)%tw.size].Flush()...)

		if tw.overflow != nil {
			overflowNodes := tw.overflow.advanceClock(time.Unix(0, tw.currentTime))

			for _, node := range overflowNodes {
				if tw.add(node.Task) == nil {
					dueNodes = append(dueNodes, node)
				}
			}
		}
	}

	return
}
