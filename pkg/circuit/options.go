package circuit

import "time"

const (
	defaultBucketTime  = time.Millisecond * 100
	defaultBucketCount = 100
	// cooling timeout is the time the breaker stay in Open before becoming HalfOpen
	defaultCoolingTime = time.Second * 5

	// detect timeout is the time interval between every detect in HalfOpen
	// defaultDetectTime = time.Millisecond * 200

	// halfopen success is the threshold when the breaker is in HalfOpen;
	// after exceeding consecutively this times, it will change its State from HalfOpen to Closed;
	defaultHalfOpenSuccesses = 2
)

// Options ...
type Options struct {
	BucketCount    int                 // 时间窗桶个数
	BucketTime     time.Duration       // 时间窗口大小
	CoolingTime    time.Duration       // used in open state
	DetectTime     time.Duration       // used in half-open state
	Trip           TripFunc            //
	OnStateChanged StateChangedHandler // 状态发生变化通知
}

func (o *Options) check() {
	if o.BucketCount <= 0 {
		o.BucketCount = defaultBucketCount
	}
	if o.BucketTime <= 0 {
		o.BucketTime = defaultBucketTime
	}
	if o.CoolingTime <= 0 {
		o.CoolingTime = defaultCoolingTime
	}
	if o.DetectTime <= 0 {
		o.DetectTime = defaultCoolingTime
	}
	if o.Trip == nil {
		o.Trip = ThresholdTripFunc(100)
	}
}
