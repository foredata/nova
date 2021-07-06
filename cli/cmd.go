package cli

import "fmt"

func NewCmd(name string, usage string, broadcast bool, fn interface{}) Command {
	action, flags := toAction(fn)

	cmd := &command{name: name, usage: usage, broadcast: broadcast, action: action, flags: flags}
	return cmd
}

// Command 可执行命令，以树形形式组织，只有叶子节点的命令才可以执行Action
type Command interface {
	Name() string                         // 命令名
	Usage() string                        // 描述说明
	Broadcast() bool                      // 是否需要广播
	Flags() []*Flag                       //
	AddFlags(f ...*Flag)                  //
	Subs() []Command                      // 子命令
	Add(sub Command)                      // 添加子命令
	Get(name string) Command              // 通过名字查询sub command
	Run(ctx Context) (interface{}, error) // 执行command
}

type Action = func(ctx Context) (interface{}, error)

type command struct {
	name      string
	usage     string
	broadcast bool
	action    Action
	flags     []*Flag
	subs      []Command // 子命令
}

func (c *command) Name() string {
	return c.name
}

func (c *command) Usage() string {
	return c.usage
}

func (c *command) Broadcast() bool {
	return c.broadcast
}

func (c *command) Flags() []*Flag {
	return c.flags
}

func (c *command) AddFlags(f ...*Flag) {
	c.flags = append(c.flags, f...)
}

func (c *command) Subs() []Command {
	return c.subs
}

func (c *command) Add(sub Command) {
	c.subs = append(c.subs, sub)
}

func (c *command) Get(name string) Command {
	for _, sub := range c.subs {
		if sub.Name() == name {
			return sub
		}
	}

	return nil
}

func (c *command) Run(ctx Context) (interface{}, error) {
	if c.action != nil {
		return c.action(ctx)
	} else if len(c.subs) == 0 {
		// 叶子节点需要提供action
		return nil, fmt.Errorf("cli: no action, %+v", c.name)
	} else {
		// 非叶子节点
		return nil, nil
	}
}
