package saga

type RunOptions struct {
	txId int64
}

type RunOption func(o *RunOptions)

func newRunOptions(opts ...RunOption) *RunOptions {
	res := &RunOptions{}
	for _, fn := range opts {
		fn(res)
	}

	return res
}

type NewOptions struct {
	mgr Manager
}

type NewOption func(o *NewOptions)

func newNewOptions(opts ...NewOption) *NewOptions {
	res := &NewOptions{}
	for _, fn := range opts {
		fn(res)
	}

	return res
}
