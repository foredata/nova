package cli

import "context"

// Context 执行Command时上下文
type Context interface {
	context.Context
	// 获取Args
	NArg() int
	Args() []string
	Arg(index int) string
	Tail(index int) []string
	// 获取Flag
	NFlag() int
	Flag(key string) []string
}

type ccontext struct {
	context.Context
	flags map[string][]string
	args  []string
}

func (c *ccontext) NArg() int {
	return len(c.args)
}

func (c *ccontext) Args() []string {
	return c.args
}

func (c *ccontext) Arg(index int) string {
	return c.args[index]
}

func (c *ccontext) Tail(index int) []string {
	if index < len(c.args) {
		return c.args[index:]
	}

	return nil
}

func (c *ccontext) NFlag() int {
	return len(c.flags)
}

func (c *ccontext) Flag(key string) []string {
	return c.flags[key]
}
