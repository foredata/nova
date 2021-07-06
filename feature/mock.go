package feature

import "context"

// NewMockClient .
func NewMockClient() Client {
	return &mockClient{}
}

// TODO: setup mock config and caculate evaluate result
// https://github.com/Knetic/govaluate
type mockClient struct {
}

func (c *mockClient) IsToggled(ctx context.Context, key string, attributes Attributes, defaultVal bool) (bool, error) {
	return defaultVal, nil
}

func (c *mockClient) Evaluate(ctx context.Context, key string, attributes Attributes, defaultVal string) (string, error) {
	return defaultVal, nil
}
