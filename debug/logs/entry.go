package logs

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Entry 一条输出日志
type Entry struct {
	sync.RWMutex
	Logger    Logger               // 日志Owner
	Level     Level                // 日志级别
	Text      string               // 日志信息
	Tags      SortedMap            // Tags,初始化时设置的标签信息,比如env,host等
	Fields    []Field              // 附加字段,无序,k=v格式整体输出
	Time      time.Time            // 时间戳
	Context   context.Context      // 上下文,通常用于填充Fields
	Host      string               // 配置host
	Path      string               // 文件全路径,包含文件名
	File      string               // 文件名
	Line      int                  // 行号
	Method    string               // 方法名
	CallDepth int                  // 需要忽略的堆栈
	outputs   map[Formatter][]byte // 相同的Formater只会构建一次
	refs      int32                // 引用计数,当为0时,会放到缓存中
}

var gEntryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{}
	},
}

// NewEntry 创建Entry
func NewEntry(logger Logger) *Entry {
	e := gEntryPool.Get().(*Entry)
	e.Logger = logger
	e.Time = time.Now()
	e.CallDepth = DefaultCallDepth
	e.outputs = make(map[Formatter][]byte)
	e.refs = 1
	return e
}

// Obtain 增加引用计数
func (e *Entry) Obtain() {
	atomic.AddInt32(&e.refs, 1)
}

// Free 当引用计数为0
func (e *Entry) Free() {
	if atomic.AddInt32(&e.refs, -1) <= 0 {
		gEntryPool.Put(e)
	}
}
