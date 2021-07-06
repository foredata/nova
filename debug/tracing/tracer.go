package tracing

type Tracer interface {
	// StartSpan starts a span with the given operation name and options.
	StartSpan(name string, opts ...StartSpanOption) Span

	// Extract extracts a span context from a given carrier. Note that baggage item
	// keys will always be lower-cased to maintain consistency. It is impossible to
	// maintain the original casing due to MIME header canonicalization standards.
	Extract(carrier interface{}) (SpanContext, error)

	// Inject injects a span context into the given carrier.
	Inject(context SpanContext, carrier interface{}) error

	// Stop stops the tracer. Calls to Stop should be idempotent.
	Stop()
}

// Span represents a chunk of computation time. Spans have names, durations,
// timestamps and other metadata. A Tracer is used to create hierarchies of
// spans in a request, buffer and submit them to the server.
type Span interface {
	// Context() yields the SpanContext for this Span. Note that the return
	// value of Context() is still valid after a call to Span.Finish(), as is
	// a call to Span.Context() after a call to Span.Finish().
	Context() SpanContext
	// SetTag sets a key/value pair as metadata on the span.
	SetTag(key string, value interface{})
	// SetBaggageItem sets a new baggage item at the given key. The baggage
	// item should propagate to all descendant spans, both in- and cross-process.
	SetBaggageItem(key, value string)
	// Finish finishes the current span with the given options. Finish calls should be idempotent.
	Finish()
}

// SpanContext represents a span state that can propagate to descendant spans
// and across process boundaries(e.g., a <trace_id, span_id, sampled> tuple).
// It contains all the information needed to
// spawn a direct descendant of the span that it belongs to. It can be used
// to create distributed tracing by propagating it using the provided interfaces.
type SpanContext interface {
	// ForeachBaggageItem provides an iterator over the key/value pairs set as
	// baggage within this context. Iteration stops when the handler returns
	// false.
	// ForeachBaggageItem(handler func(k, v string) bool)
}
