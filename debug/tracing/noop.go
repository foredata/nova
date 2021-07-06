package tracing

type noopTracer struct{}

func (noopTracer) StartSpan(name string, opts ...StartSpanOption) Span {
	return noopSpan{}
}

func (noopTracer) Extract(carrier interface{}) (SpanContext, error) {
	return noopSpanContext{}, nil
}

func (noopTracer) Inject(ctx SpanContext, carrier interface{}) error {
	return nil
}

func (noopTracer) Stop() {
}

type noopSpan struct{}

func (n noopSpan) Context() SpanContext {
	return noopSpanContext{}
}

func (n noopSpan) SetTag(key string, value interface{}) {}
func (n noopSpan) SetBaggageItem(key, value string)     {}

func (n noopSpan) Finish() {}

//
type noopSpanContext struct{}

func (noopSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}
