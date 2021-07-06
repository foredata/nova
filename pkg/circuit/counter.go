package circuit

import (
	"sync"
	"time"
)

// NewCounter ...
func NewCounter(bucketCount int) Counter {
	c := &counter{}
	c.buckets = make([]*bucket, bucketCount)
	for i := 0; i < bucketCount; i++ {
		c.buckets = append(c.buckets, &bucket{})
	}

	return c
}

type bucket struct {
	success int64
	failure int64
	timeout int64
}

func (b *bucket) Reset() {
	b.success = 0
	b.failure = 0
	b.timeout = 0
}

// counter 基于滑动窗口的方式统计数据
type counter struct {
	mux            sync.RWMutex
	buckets        []*bucket // window bucket,ring buffer
	offset         int       // ring buffer offset
	counts         Counts    // 统计计数
	conseStartTime int64     // 起始时间
}

func (c *counter) Counts() Counts {
	c.mux.RLock()
	defer c.mux.RUnlock()
	c.counts.ConseTime = time.Duration(NowUnixNano() - c.conseStartTime)
	return c.counts
}

func (c *counter) Succeed() {
	c.mux.Lock()
	b := c.getBucket()
	b.success += 1
	c.conseStartTime = 0
	c.counts.ConseErrors = 0
	c.counts.Successes += 1
	c.mux.Unlock()
}

func (c *counter) Fail() {
	c.mux.Lock()
	b := c.getBucket()
	b.failure += 1
	c.counts.Failures += 1
	c.counts.ConseErrors += 1
	if c.conseStartTime == 0 {
		c.conseStartTime = NowUnixNano()
	}
	c.mux.Unlock()
}

func (c *counter) Timeout() {
	c.mux.Lock()
	b := c.getBucket()
	b.timeout += 1
	c.counts.Timeouts += 1
	c.counts.ConseErrors += 1
	if c.conseStartTime == 0 {
		c.conseStartTime = NowUnixNano()
	}
	c.mux.Unlock()
}

func (c *counter) Reset() {
	c.mux.Lock()
	c.counts.reset()
	c.conseStartTime = 0
	c.offset = 0
	for _, b := range c.buckets {
		b.Reset()
	}
	c.mux.Unlock()
}

func (c *counter) Tick() {
	c.mux.Lock()
	c.offset = c.offset + 1
	if c.offset >= len(c.buckets) {
		c.offset = 0
	}

	old := c.getBucket()
	counts := &c.counts
	counts.Successes -= old.success
	counts.Failures -= old.failure
	counts.Timeouts -= old.timeout
	old.Reset()
	c.mux.Unlock()
}

func (c *counter) getBucket() *bucket {
	return c.buckets[c.offset]
}
