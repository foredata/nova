package job

import "time"

type RunOptions struct {
	// 运行等待时间,超过此时间会返回ErrJobTimeout,业务可忽略此错误,0则立即返回
	WaitTime time.Duration
}

type RunOption func(o *RunOptions)

func WithWaitTime(wt time.Duration) RunOption {
	return func(o *RunOptions) {
		o.WaitTime = wt
	}
}
