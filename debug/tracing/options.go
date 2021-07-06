package tracing

type StartSpanOptions struct {
	Parent SpanContext
	Tags   map[string]interface{}
}

type StartSpanOption func(o *StartSpanOptions)

// WithParent 通过Span设置parent
func WithParent(p Span) StartSpanOption {
	return func(o *StartSpanOptions) {
		o.Parent = p.Context
	}
}

// WithParentContext 通过SpanContext设置Parent
func WithParentContext(p SpanContext) StartSpanOption {
	return func(o *StartSpanOptions) {
		o.Parent = p
	}
}

// WithTag 设置tag
func WithTag(key string, value interface{}) StartSpanOption {
	return func(o *StartSpanOptions) {
		if o.Tags == nil {
			o.Tags = make(map[string]interface{})
		}
		o.Tags[key] = value
	}
}

// WithTags 设置tags
func WithTags(tags map[string]interface{}) StartSpanOption {
	return func(o *StartSpanOptions) {
		o.Tags = tags
	}
}
