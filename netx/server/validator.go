package server

import "context"

// Validator is the interface that wraps the Validate function.
type Validator interface {
	Validate(ctx context.Context, msg interface{}) error
}
