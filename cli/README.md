# cli(Command Line Interface)

cli控制台模块，区别于cobra和urfave/cli这种用于命令行工具的库，这里通常用于执行一些后台Admin命令，比如手动加载配置，执行特定任务等。

## 特点：

- 支持传统命令行格式：如: kubectl get -f pod.yaml -o json
- 支持多级command
- 支持广播模式：比如当在集群中手动执行配置加载时，需要所有机器都执行，而不是某台机器
- 触发方式：可以是控制台，Bot，Admin中台

## 使用方法：

```go
// 定义Command参数
type demoCmd struct {
	Name    string   `flag:"name|n, required,hidden,usage='demo,asdfsdf'"` // 可选参数-n或-name
	File    string   // 普通参数，按照顺序解析,只有最后一个可以是数组
	Param1  string   //
	Remains []string // 剩余参数
}

// 定义Command处理函数
func onDemo(cmd *demoCmd) error {
	fmt.Println("process Demo", cmd.Name, cmd.File, cmd.Param1, cmd.Remains)
	return nil
}

func TestDemo(t *testing.T) {
    // 注册Command
    // 如果需要Broadcast Command,需要实现Pubsub接口,并覆盖默认App
    cli.Register(cli.NewCmd("demo", "demo", false, onDemo))

    // 触发执行Command
    cli.Exec(context.Background(), []string{"demo", "-n=name", "filename", "param1", "remain 1", "remain 2"})
}
```

## other library

- [cobra](https://github.com/spf13/cobra)
- [urfave/cli](https://github.com/urfave/cli)