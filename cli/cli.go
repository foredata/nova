package cli

import (
	"context"
	"fmt"

	"github.com/foredata/nova/pkg/flags"
)

var defaultApp = NewApp(nil)

// SetDefault 设置默认
func SetDefault(app App) {
	defaultApp = app
}

// Register 注册Command
func Register(cmd Command) {
	defaultApp.Register(cmd)
}

// Exec 由外部触发执行Command
func Exec(ctx context.Context, args []string) (interface{}, error) {
	return defaultApp.Exec(ctx, args)
}

// App .
type App interface {
	// Commands 返回所有注册的Commands
	Commands() []Command
	// Register 注册Command,重复会panic
	Register(cmd Command)
	// Exec 执行Command
	Exec(ctx context.Context, args []string) (interface{}, error)
}

// Config 配置信息
type Config struct {
	PubSub PubSub
}

// NewApp 新建App
func NewApp(conf *Config) App {
	a := &app{
		dict: make(map[string]Command),
	}
	if conf != nil && conf.PubSub != nil {
		a.pubsub = conf.PubSub
	} else {
		a.pubsub = &localPubsub{}
	}
	a.pubsub.Subscribe(a.onSubscribe)
	return a
}

type app struct {
	cmds   []Command          //
	dict   map[string]Command //
	pubsub PubSub             //
}

func (a *app) Commands() []Command {
	return a.cmds
}

func (a *app) Register(cmd Command) {
	name := cmd.Name()
	if len(name) == 0 {
		panic(fmt.Errorf("cli: invalid command name"))
	}

	if a.dict[name] != nil {
		panic(fmt.Errorf("cli: duplicate command, %+v", name))
	}
	a.cmds = append(a.cmds, cmd)
	a.dict[name] = cmd
}

func (a *app) Exec(ctx context.Context, args []string) (interface{}, error) {
	return a.doExec(ctx, args, false)
}

func (a *app) onSubscribe(ctx context.Context, args []string) (interface{}, error) {
	return a.doExec(ctx, args, true)
}

func (a *app) doExec(ctx context.Context, args []string, ignoreBroadcast bool) (interface{}, error) {
	params, options, err := flags.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("cli: parse args fail, %w", err)
	}

	if len(params) == 0 {
		return nil, fmt.Errorf("cli: invalid args, no command, args=%+v", args)
	}

	name := params[0]
	cmd := a.dict[name]
	params = params[1:]
	if cmd == nil {
		return nil, fmt.Errorf("cli: not found command, name=%+v", name)
	}

	for {
		subs := cmd.Subs()
		if len(subs) == 0 {
			break
		}

		if len(params) == 0 {
			return nil, fmt.Errorf("cli: no sub command, parent=%+v, args=%+v", cmd.Name(), args)
		}
		name := params[0]
		params = params[1:]
		sub := cmd.Get(name)
		if sub == nil {
			return nil, fmt.Errorf("cli: not found sub command, parent=%+v, sub=%+v", cmd.Name(), name)
		}
		cmd = sub
	}

	if !ignoreBroadcast && cmd.Broadcast() {
		// 需要广播执行,先publish
		a.pubsub.Publish(ctx, args)
		return nil, nil
	}

	cctx := &ccontext{Context: ctx, flags: options, args: params}
	return cmd.Run(cctx)
}

// TODO: 格式化输出 help
func (a *app) doHelp() {

}
