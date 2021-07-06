package cli_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/foredata/nova/cli"
)

func TestCli(t *testing.T) {
	cli.Register(cli.NewCmd("demo", "demo", false, onDemo))

	res, err := cli.Exec(context.Background(), []string{"demo", "-n=name", "filename", "param1", "remain 1", "remain 2"})
	t.Log(res)
	t.Log(err)
}

type demoCmd struct {
	Name    string   `flag:"name|n, required,hidden,usage='demo,asdfsdf'"` // 可选参数-n或-name
	File    string   // 普通参数，按照顺序解析,只有最后一个可以是数组
	Param1  string   //
	Remains []string // 剩余参数
}

func onDemo(cmd *demoCmd) error {
	fmt.Println("process Demo", cmd.Name, cmd.File, cmd.Param1, cmd.Remains)
	return nil
}
