# timing

高效的定时器管理，与官方的timer区别
- 实现算法：这里采用Hierarchical Time Wheel，官方采用小顶堆
- 调度方式：这里只有1个协程管理定时器，1个协程执行过期回调
- 执行方式：这里采用callback的方式，官方采用channel的方式

## 使用方式

```go
func TestTimer(t *testing.T) {
	fmt.Printf("start %+v\n", time.Now())

	timing.NewDelayer(time.Second*2, func(data interface{}) {
		fmt.Printf("delay %+v, 2\n", time.Now())
	}, nil)

	timing.NewDelayer(time.Second*4, func(data interface{}) {
		fmt.Printf("delay %+v, 4\n", time.Now())
	}, nil)

	i := 0
	timing.NewTicker(time.Second, func(data interface{}) {
		i++
		fmt.Printf("tick  %+v, %+v\n", time.Now(), i)
	}, nil)

	time.Sleep(time.Second * 10)
}
```