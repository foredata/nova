package tracing

var tracer Tracer = &noopTracer{}

func SetDefault(t Tracer) {
	tracer.Stop()
	tracer = t
}

func Default() Tracer {
	return tracer
}

func StartSpan(name string, opts ...StartSpanOption) Span {
	return tracer.StartSpan(name, opts...)
}

func Extract(carrier interface{}) (SpanContext, error) {
	return tracer.Extract(carrier)
}

func Inject(ctx SpanContext, carrier interface{}) error {
	return tracer.Inject(ctx, carrier)
}

func Stop() {
	tracer.Stop()
}
