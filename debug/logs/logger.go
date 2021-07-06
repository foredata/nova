package logs

import (
	"fmt"
	"path/filepath"
)

var (
	DefaultFormatter = MustNewTextFormatter(defaultTextLayout)
)

const (
	DefaultMax       = 10000 // 默认日志最大条数
	DefaultCallDepth = 3     // 堆栈深度,忽略log相关的堆栈
)

// Channel 代表日志输出通路
type Channel interface {
	IsEnable(lv Level) bool
	Level() Level
	SetLevel(lv Level)
	Name() string
	Open() error
	Close() error
	Write(msg *Entry)
}

// BatchChannel 以batch形式发送,有需要的可以
type BatchChannel interface {
	Channel
	WriteBatch(msg []*Entry)
}

// Filter 在每天日志写入Channel前统一预处理,若返回错误则忽略该条日志
// 可用于通过Context添加Field,对某些Field加密等处理
type Filter func(*Entry) error

// Formatter 用于格式化Entry
// 相同的Formatter对每个Entry只会格式化一次
type Formatter interface {
	Name() string                      //
	Format(msg *Entry) ([]byte, error) // 格式化输出
}

// Sampler 日志采样,对于高频的日志可以限制发送频率
type Sampler interface {
	Check(msg *Entry) bool
}

// Config 配置信息
type Config struct {
	Channels      []Channel // 日志输出通路,至少1个,默认Console
	Filters       []Filter  // 过滤函数
	Tags          SortedMap // 全局Fields,比如env,cluster,psm,host等
	Level         Level     // 日志级别,默认Info
	LogMax        int       // 最大缓存日志数
	DisableCaller bool      // 是否关闭Caller,若为true则获取不到文件名等信息
	Async         bool      // 是否异步,默认同步
}

func (c *Config) AddChannels(channels ...Channel) {
	c.Channels = append(c.Channels, channels...)
}

// AddTags 添加Tags
func (c *Config) AddTags(tags map[string]string) {
	c.Tags.Fill(tags)
}

// NewConfig 创建配置
func NewConfig() *Config {
	return &Config{
		Level:  TraceLevel,
		LogMax: DefaultMax,
	}
}

// Logger 日志系统接口(structured, leveled logging)
// 1:支持同步模式和异步模式
// 	在Debug模式下,通常使用同步模式,因为可以保证console与fmt顺序一致
// 	正式环境下可以使用异步模式,保证日志不会影响服务质量,不能保证日志不丢失
// 2:配置信息
//	大部分配置信息是不能动态更新的,异步处理时并没有枷锁,比如Channel,Filter,Tags
//	部分简单配置是可以动态更新的,比如Level降级,可用于临时调试
// 3:关于Context
//	通过Context可以透传RequestID,LogID，UID等信息,Log第一个参数都强制要求传入Ctx,但可以为nil
// 	Logger本身并不知道如何处理Ctx,因此需要初始化时手动添加Filter用来解析Context
type Logger interface {
	IsEnable(lv Level) bool
	SetLevel(name string, lv Level)
	Start()
	Stop()
	Write(e *Entry)
	Log(lv Level, msg string, fields ...Field)
	Logf(lv Level, format string, args ...interface{})
}

// NewLogger 创建默认的Logger
func NewLogger(config *Config) Logger {
	l := &logger{Config: config}
	if config.Async {
		l.channels = append(l.channels, NewAsyncChannel(l.channels, config.LogMax))
		l.Start()
	} else {
		l.channels = config.Channels
	}

	return l
}

type logger struct {
	*Config
	channels []Channel
}

func (l *logger) getChannel(name string) Channel {
	for _, c := range l.Channels {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

func (l *logger) IsEnable(lv Level) bool {
	return lv <= l.Level
}

func (l *logger) SetLevel(name string, lv Level) {
	if name == "" {
		l.Level = lv
	} else if c := l.getChannel(name); c != nil {
		c.SetLevel(lv)
	}
}

// Start run async logger
func (l *logger) Start() {
	for _, c := range l.channels {
		_ = c.Open()
	}
}

// Stop stop async logger
func (l *logger) Stop() {
	for _, c := range l.channels {
		c.Close()
	}
}

func (l *logger) Write(e *Entry) {
	if !l.DisableCaller {
		f := getFrame(e.CallDepth)
		e.Path = f.File
		e.File = filepath.Base(f.File)
		e.Line = f.Line
		e.Method = getFuncName(f.Function)
	}
	e.Tags = l.Tags

	for _, f := range l.Filters {
		if err := f(e); err != nil {
			return
		}
	}

	for _, c := range l.channels {
		if c.IsEnable(e.Level) {
			c.Write(e)
		}
	}
}

func (l *logger) Log(lv Level, msg string, fields ...Field) {
	if l.IsEnable(lv) {
		e := NewEntry(l)
		e.Level = lv
		e.Text = msg
		e.Fields = fields
		l.Write(e)
	}
}

func (l *logger) Logf(lv Level, format string, args ...interface{}) {
	if l.IsEnable(lv) {
		text := fmt.Sprintf(format, args...)
		e := NewEntry(l)
		e.Level = lv
		e.Text = text
		l.Write(e)
	}
}
