package htw_test

import (
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTimingWheel_AddAndExpire(t *testing.T) {
	start := time.Unix(100, 0)

	tw := htw.NewTimingWheel[string](10*time.Millisecond, start, 10)

	require.False(
		t,
		tw.Add(htw.NewTask(uuid.New(), start.Add(2*time.Millisecond), "too-early")),
		"Tasks falling into the current/past tick should not be added",
	)

	validTask := htw.NewTask(uuid.New(), start.Add(35*time.Millisecond), "in-range")
	require.True(t, tw.Add(validTask))

	require.Empty(t, tw.AdvanceClock(start.Add(20*time.Millisecond)), "No tasks should be due yet")

	due := tw.AdvanceClock(start.Add(40 * time.Millisecond))
	require.Len(t, due, 1)
	require.Equal(t, validTask.Id, due[0].Id)
}

func TestTimingWheel_OverflowAndCascade(t *testing.T) {
	start := time.Unix(100, 0)

	tw := htw.NewTimingWheel[string](10*time.Millisecond, start, 10)

	overflowTask := htw.NewTask(uuid.New(), start.Add(250*time.Millisecond), "overflow-payload")
	require.True(t, tw.Add(overflowTask), "Should successfully route to overflow wheel")

	due := tw.AdvanceClock(start.Add(150 * time.Millisecond))
	require.Empty(t, due, "Task is slated for +250ms, shouldn't be due at +150ms")

	due = tw.AdvanceClock(start.Add(260 * time.Millisecond))
	require.Len(t, due, 1)
	require.Equal(t, overflowTask.Id, due[0].Id)
}

func TestTimingWheel_PreciseBucketRouting(t *testing.T) {
	start := time.Unix(0, 0)

	tw := htw.NewTimingWheel[string](1*time.Second, start, 4)

	taskAt1s := htw.NewTask(uuid.New(), start.Add(1*time.Second), "t1")
	taskAt3s := htw.NewTask(uuid.New(), start.Add(3*time.Second), "t3")

	require.True(t, tw.Add(taskAt1s))
	require.True(t, tw.Add(taskAt3s))

	due := tw.AdvanceClock(start.Add(1500 * time.Millisecond))
	require.Len(t, due, 1)
	require.Equal(t, "t1", due[0].Value)

	require.Empty(t, tw.AdvanceClock(start.Add(2500*time.Millisecond)))

	due = tw.AdvanceClock(start.Add(3500 * time.Millisecond))
	require.Len(t, due, 1)
	require.Equal(t, "t3", due[0].Value)
}
