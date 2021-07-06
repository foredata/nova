package netx

// BaseFilter 默认实现Filter接口,便于使用者无须实现不必要的接口
type BaseFilter struct {
}

// HandleRead ...
func (f *BaseFilter) HandleRead(ctx FilterCtx) error {
	return nil
}

// HandleWrite ...
func (f *BaseFilter) HandleWrite(ctx FilterCtx) error {
	return nil
}

// HandleOpen ...
func (f *BaseFilter) HandleOpen(ctx FilterCtx) error {
	return nil
}

// HandleClose ...
func (f *BaseFilter) HandleClose(ctx FilterCtx) error {
	return nil
}

// HandleError ...
func (f *BaseFilter) HandleError(ctx FilterCtx) error {
	return nil
}
