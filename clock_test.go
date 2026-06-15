package htw_test

import (
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/stretchr/testify/require"
)

func TestTickerClock_InterfaceImplementation(t *testing.T) {
	var _ htw.Clock = (*htw.TickerClock)(nil)

	clock := htw.NewTickerClock(1 * time.Millisecond)
	defer clock.Stop()

	require.NotNil(t, clock.TickChan())
	require.WithinDuration(t, time.Now(), clock.Now(), 1*time.Second)
}
