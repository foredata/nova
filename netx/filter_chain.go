package netx

// NewFilterChain 创建FitlerChain
func NewFilterChain() FilterChain {
	return &filterChain{}
}

// 分为InBound和OutBound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type filterChain struct {
	filters []Filter
}

func (fc *filterChain) Len() int {
	return len(fc.filters)
}

func (fc *filterChain) Front() Filter {
	if fc.Len() > 0 {
		return fc.filters[0]
	}

	return nil
}

func (fc *filterChain) Back() Filter {
	if fc.Len() > 0 {
		return fc.filters[fc.Len()-1]
	}

	return nil
}

func (fc *filterChain) Get(index int) Filter {
	return fc.filters[index]
}

func (fc *filterChain) Index(name string) int {
	for index, filter := range fc.filters {
		if filter.Name() == name {
			return index
		}
	}

	return -1
}

func (fc *filterChain) AddFirst(filters ...Filter) {
	filters = append(filters, fc.filters[1:]...)
	fc.filters = append(fc.filters[0:0], filters...)
}

func (fc *filterChain) AddLast(filters ...Filter) {
	fc.filters = append(fc.filters, filters...)
}

func (fc *filterChain) HandleOpen(conn Conn) {
	ctx := newFilterCtx(fc.filters, conn, true, doOpen)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *filterChain) HandleClose(conn Conn) {
	ctx := newFilterCtx(fc.filters, conn, true, doClose)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *filterChain) HandleRead(conn Conn, msg interface{}) {
	ctx := newFilterCtx(fc.filters, conn, true, doRead)
	ctx.SetData(msg)
	if err := ctx.Call(); err != nil {
		fc.HandleError(conn, err)
	}
}

func (fc *filterChain) HandleWrite(conn Conn, msg interface{}) error {
	fctx := newFilterCtx(fc.filters, conn, false, doWrite)
	fctx.SetData(msg)
	err := fctx.Call()
	if err != nil {
		fc.HandleError(conn, err)
	} else if !fctx.IsAbort() {
		if p, ok := fctx.Data().(WriterTo); ok {
			return fctx.Conn().Write(p)
		}
	}

	return err
}

func (fc *filterChain) HandleError(conn Conn, err error) {
	ctx := newFilterCtx(fc.filters, conn, false, doError)
	ctx.SetError(err)
	// 忽略错误,防止无限递归
	_ = ctx.Call()
}

func doOpen(f Filter, ctx FilterCtx) error {
	return f.HandleOpen(ctx)
}

func doClose(f Filter, ctx FilterCtx) error {
	return f.HandleClose(ctx)
}

func doRead(f Filter, ctx FilterCtx) error {
	return f.HandleRead(ctx)
}

func doWrite(f Filter, ctx FilterCtx) error {
	return f.HandleWrite(ctx)
}

func doError(f Filter, ctx FilterCtx) error {
	return f.HandleError(ctx)
}
