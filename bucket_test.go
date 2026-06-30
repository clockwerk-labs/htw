package htw_test

import (
	"sync"
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/stretchr/testify/require"
)

func TestBucket_Sequential(t *testing.T) {
	bucket := htw.NewBucket[string]()

	require.Nil(t, bucket.Flush())

	task1 := htw.NewTask(time.Now().Add(1*time.Hour), "first-task")
	task2 := htw.NewTask(time.Now().Add(1*time.Hour), "second-task")

	node1 := bucket.Add(task1)
	node2 := bucket.Add(task2)

	evicted := bucket.Flush()

	require.Len(t, evicted, 2)
	require.Equal(t, []*htw.Node[string]{node1, node2}, evicted)
	require.Nil(t, bucket.Flush())
}

func TestBucket_Concurrent(t *testing.T) {
	bucket := htw.NewBucket[string]()

	numGoroutines := 10
	tasksPerGoroutine := 100
	exp := time.Now().Add(5 * time.Minute)

	var addWg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		addWg.Go(func() {
			for j := 0; j < tasksPerGoroutine; j++ {
				bucket.Add(htw.NewTask(exp, "payload"))
			}
		})
	}

	var flushedNodes []*htw.Node[string]
	var flushWg sync.WaitGroup

	flushWg.Go(func() {
		addWg.Wait()
		flushedNodes = append(flushedNodes, bucket.Flush()...)
	})

	flushWg.Wait()

	require.Len(t, flushedNodes, numGoroutines*tasksPerGoroutine, "Data loss detected!")
}
