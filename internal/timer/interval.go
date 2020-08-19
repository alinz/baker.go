package timer

import (
	"context"
	"time"
)

// Interval calls given fn for every duration
// calling cancel will cancel the timer and propagate the cancel into given func
func Interval(ms time.Duration, fn func(ctx context.Context)) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-time.After(ms):
				go fn(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel
}
