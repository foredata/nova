package feature

import "context"

// defaultClient 全局默认client,默认为mock client，使用时通常需要替换
var defaultClient = NewMockClient()

// SetDefault 设置全局默认Client
func SetDefault(c Client) {
	defaultClient = c
}

// IsToggled .
func IsToggled(ctx context.Context, key string, attributes Attributes, defaultVal bool) (bool, error) {
	return defaultClient.IsToggled(ctx, key, attributes, defaultVal)
}

// Evaluate .
func Evaluate(ctx context.Context, key string, attributes Attributes, defaultVal string) (string, error) {
	return defaultClient.Evaluate(ctx, key, attributes, defaultVal)
}
