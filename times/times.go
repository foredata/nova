package times

import (
	"time"

	"github.com/foredata/nova/times/clock"
)

// defaultClock 默认clock
var defaultClock = clock.New()

// SetClock 替换默认clock
func SetClock(c clock.Clock) {
	defaultClock = c
}

// Now .
func Now() time.Time {
	return defaultClock.Now()
}

// NowUnix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
// Unix-like operating systems often record time as a 32-bit
// count of seconds, but since the method here returns a 64-bit
// value it is valid for billions of years into the past or future.
func NowUnix() int64 {
	return defaultClock.Now().Unix()
}

// Since .
func Since(t time.Time) time.Duration {
	return defaultClock.Since(t)
}

// Sleep .
func Sleep(d time.Duration) {
	defaultClock.Sleep(d)
}

// Start 启动定时器,只会触发1次
func Start() uint64 {
	return 0
}
