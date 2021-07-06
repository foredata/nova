package saga

import "context"

type ctxKey struct{}
type ctxParam map[string]interface{}

func Get(ctx context.Context, key string) interface{} {
	return nil
}

// Set 设置透传参数
func Set(ctx context.Context, key string, value interface{}) context.Context {
	return ctx
}

func GetTxID(ctx context.Context) int64 {
	return 0
}

func NewCtx(ctx context.Context, tx int64) context.Context {
	return ctx
}
