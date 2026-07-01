package htw_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/clockwerk-labs/htw"
	"github.com/stretchr/testify/require"
)

func TestRegistry_BasicOperations(t *testing.T) {
	reg := htw.NewRegistry[int, string](4)

	reg.Add(1, "Task1")
	reg.Add(2, "Task2")

	val, found := reg.Get(1)
	require.True(t, found, "Key 1 should be found")
	require.Equal(t, "Task1", val)

	removedVal, removed := reg.Remove(1)
	require.True(t, removed, "Key 1 should be successfully removed")
	require.Equal(t, "Task1", removedVal)

	_, found = reg.Get(1)
	require.False(t, found, "Key 1 should no longer exist")

	reg.Clear()
	_, found = reg.Get(2)
	require.False(t, found, "Registry should be empty after Clear")
}

func TestRegistry_GetAll_And_DeadlockImmunity(t *testing.T) {
	reg := htw.NewRegistry[string, string](8)
	reg.Add("k1", "v1")
	reg.Add("k2", "v2")
	reg.Add("k3", "v3")

	count := 0
	for range reg.GetAll() {
		count++
		reg.Add(fmt.Sprintf("new_key_%d", count), "new_val")
	}

	require.Equal(t, 3, count, "Should only iterate over the 3 initial snapshot elements")
}

func TestRegistry_Concurrency(t *testing.T) {
	reg := htw.NewRegistry[int, int](16)
	var wg sync.WaitGroup

	numGoroutines := 50
	operationsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Go(func() {
			for j := 0; j < operationsPerGoroutine; j++ {
				key := i*operationsPerGoroutine + j
				reg.Add(key, j)
			}
		})
	}

	for i := 0; i < 5; i++ {
		wg.Go(func() {
			for val := range reg.GetAll() {
				_ = val
			}
		})
	}

	wg.Wait()

	totalExpectedItems := numGoroutines * operationsPerGoroutine
	actualItems := 0
	for range reg.GetAll() {
		actualItems++
	}

	require.Equal(t, totalExpectedItems, actualItems)
}
