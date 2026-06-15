package htw

import "time"

type (
	Clock interface {
		Now() time.Time
		TickChan() <-chan time.Time
		Stop()
	}

	TickerClock struct {
		ticker *time.Ticker
	}
)

func NewTickerClock(tick time.Duration) *TickerClock {
	return &TickerClock{
		ticker: time.NewTicker(tick),
	}
}

func (*TickerClock) Now() time.Time {
	return time.Now()
}

func (t *TickerClock) TickChan() <-chan time.Time {
	return t.ticker.C
}

func (t *TickerClock) Stop() {
	t.ticker.Stop()
}
