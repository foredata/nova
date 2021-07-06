package domain

// HandleOptions 执行Handle可选参数
type HandleOptions struct {
	aggType AggregateType
	aggId   string
}

type HandleOption func(o *HandleOptions)

func newHandleOptions(opts ...HandleOption) *HandleOptions {
	res := &HandleOptions{}
	for _, fn := range opts {
		fn(res)
	}

	return res
}

// WithAggregate 当command未实现Aggregatable接口时,可由外部提供Aggregate相关信息
func WithAggregate(aggType AggregateType, aggId string) HandleOption {
	return func(o *HandleOptions) {
		o.aggType = aggType
		o.aggId = aggId
	}
}
