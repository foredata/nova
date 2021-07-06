package job

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/foredata/nova/pkg/contexts"
	"github.com/foredata/nova/pkg/flags"
	"github.com/foredata/nova/pkg/reflects"
	"github.com/foredata/nova/pkg/xid"
	"github.com/foredata/nova/times"
)

// some error
var (
	ErrInvalidParams = errors.New("invalid params")
	ErrJobNotFound   = errors.New("job not found")
	ErrJobDisable    = errors.New("job disable")
	ErrJobHasRunning = errors.New("job has running")
	ErrJobTimeout    = errors.New("job wait timeout")
)

// Status 任务状态
type Status string

const (
	StatusIdle    Status = "idle"    // 空闲状态
	StatusRunning Status = "running" // 运行中
	StatusFinish  Status = "finish"  // 运行完成
	StatusError   Status = "error"   // 运行失败
	StatusCancel  Status = "cancel"  // 主动取消(但任务并不一定能退出,退出后会变为finish状态)
)

func (s Status) Is(o Status) bool {
	return s == o
}

// Callback 任务回调
type Callback func(ctx context.Context, args []string) error

// Job 任务配置信息
type Job struct {
	Name     string   // 注册任务名
	Disable  bool     // 是否被禁止
	Multiple bool     // 是否允许多任务同时执行,若为false则同一时刻只会有一个任务在执行
	Callback Callback // 执行回调
}

// Info 任务执行相关信息
type Info struct {
	Status     Status            // 当前运行状态
	StartTime  time.Time         // 开始时间
	FinishTime time.Time         // 结束时间
	CancelTime time.Time         // 取消时间
	Error      string            // 退出错误信息
	Extra      map[string]string // 用户自定义数据
}

// instance 运行实例
type instance struct {
	job    *Job
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	info   Info
}

func (ins *instance) save() {
	_ = gStore.Save(ins.ctx, ins.id, &ins.info, ins.job)
}

type jobCtxKey struct{}

var gJobMap = make(map[string]*Job)
var gInsMap = make(map[string]*instance)
var gMutex sync.RWMutex
var gStore Store = newNoopStore()

// SetStore 设置全局Store
func SetStore(s Store) {
	if s == nil {
		return
	}
	gStore = s
}

// GetJobs 获取所有注册的job
func GetJobs() map[string]*Job {
	return gJobMap
}

// Register 注册任务回调,仅能在程序启动时注册
func Register(name string, cb interface{}) {
	gMutex.Lock()
	defer gMutex.Unlock()
	job := gJobMap[name]
	if job != nil && job.Callback != nil {
		panic(fmt.Errorf("duplicate job name, %+v", name))
	}

	callback := toCallback(cb)
	if job == nil {
		job = &Job{}
		gJobMap[name] = job
	}

	job.Callback = callback
}

// Run 执行任务,虽然是后台异步执行,但允许等待一段时间,因为有些任务可以快速结束,比如发生错误或任务很简单
func Run(ctx context.Context, name string, args []string, opts ...RunOption) (string, error) {
	if name == "" {
		return "", ErrInvalidParams
	}

	o := &RunOptions{}
	for _, fn := range opts {
		fn(o)
	}

	gMutex.RLock()
	job := gJobMap[name]
	if job == nil {
		gMutex.RUnlock()
		return "", ErrJobNotFound
	}

	gMutex.RUnlock()

	if job.Disable {
		return "", ErrJobDisable
	}

	if !job.Multiple {
		ok, err := gStore.IsRunning(ctx, name)
		if err != nil {
			return "", err
		}
		if ok {
			return "", ErrJobHasRunning
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	id := xid.New().String()
	ins := &instance{id: id, ctx: ctx, job: job, cancel: cancel}
	ctx = contexts.Set(ctx, &jobCtxKey{}, ins)

	if o.WaitTime == 0 {
		err := callbackWrapper(ctx, args, ins, job.Callback)()
		return id, err
	} else {
		go callbackWrapper(ctx, args, ins, job.Callback)

		// 等待一段时间再返回结果
		t := time.NewTimer(o.WaitTime)
		for {
			select {
			case <-t.C:
				return id, ErrJobTimeout
			case <-ctx.Done():
				return id, nil
			}
		}
	}
}

// Cancel 取消任务,如果任务在当前服务则直接取消,如果不在则通知其他服务
func Cancel(ctx context.Context, id string) {
	gMutex.Lock()
	ins := gInsMap[id]
	if ins != nil {
		if ins.info.Status.Is(StatusRunning) {
			ins.info.Status = StatusCancel
			ins.info.CancelTime = times.Now()
			ins.save()
			ins.cancel()
		}
	} else {
		_ = gStore.Publish(ctx, &CancelEvent{InstanceID: id})
	}

	gMutex.Unlock()
}

// IsCancel 是否被取消,用于业务逻辑判断当前任务是否被取消,便于主动退出
func IsCancel(ctx context.Context) bool {
	ins, _ := contexts.Get(ctx, jobCtxKey{}).(*instance)
	if ins == nil {
		return false
	}

	return ins.info.Status.Is(StatusCancel)
}

// UpdateExtra 更新自定义信息
func UpdateExtra(ctx context.Context, extra map[string]string) {
	ins, _ := contexts.Get(ctx, jobCtxKey{}).(*instance)
	if ins == nil {
		return
	}
	if ins.info.Extra == nil {
		ins.info.Extra = make(map[string]string)
	}
	for k, v := range extra {
		ins.info.Extra[k] = v
	}
	ins.save()
}

// GetInfo 获取相关数据
func GetInfo(ctx context.Context) *Info {
	ins, _ := contexts.Get(ctx, jobCtxKey{}).(*instance)
	if ins == nil {
		return nil
	}

	return &ins.info
}

func callbackWrapper(ctx context.Context, args []string, ins *instance, cb Callback) func() error {
	return func() error {
		ins.info.Status = StatusRunning
		ins.info.StartTime = times.Now()
		ins.save()
		err := cb(ctx, args)
		ins.info.Status = StatusFinish
		ins.info.FinishTime = times.Now()
		if err != nil {
			ins.info.Error = err.Error()
		}
		ins.save()
		gMutex.Lock()
		delete(gInsMap, ins.id)
		gMutex.Unlock()
		return err
	}
}

// toCallback 支持的函数签名有
//	1: func(ctx context.Context, args []string) error
//	2: func(ctx context.Context) error
//	3: func(ctx context.Context, msg interface{}) error
func toCallback(cb interface{}) Callback {
	switch x := cb.(type) {
	case Callback:
		return x
	case func(ctx context.Context, args []string) error:
		return x
	}

	rv := reflect.ValueOf(cb)
	rt := rv.Type()

	if rt.Kind() != reflect.Func {
		panic(fmt.Errorf("invalid job callback type, %+v", rt.Kind()))
	}

	if rt.NumIn() < 1 || rt.NumIn() > 2 || !reflects.IsContext(rt.In(0)) {
		panic(fmt.Errorf("invalid job callback signature, input args fail"))
	}

	if rt.NumOut() != 1 || !reflects.IsError(rt.Out(0)) {
		panic(fmt.Errorf("invalid job callback signature"))
	}

	switch {
	case rt.NumIn() == 1:
		// func(ctx context.Context) error
		return func(ctx context.Context, args []string) error {
			in := []reflect.Value{reflect.ValueOf(ctx)}
			out := rv.Call(in)
			if !out[0].IsNil() {
				return out[0].Interface().(error)
			}

			return nil
		}
	case rt.NumIn() == 2:
		// func(ctx context.Context, msg interface{}) error
		if !isMessage(rt.In(1)) {
			panic(fmt.Errorf("invalid job callback signature"))
		}

		return func(ctx context.Context, args []string) error {
			msg := reflect.New(rt.In(1).Elem())
			if err := flags.Bind(msg.Interface(), args); err != nil {
				return fmt.Errorf("bind job params fail, %+v", err)
			}

			in := []reflect.Value{reflect.ValueOf(ctx), msg}
			out := rv.Call(in)
			if !out[0].IsNil() {
				return out[0].Interface().(error)
			}

			return nil
		}
	default:
		panic(fmt.Errorf("not support job callback"))
	}
}

// 用于粗略检测函数原型中参数是否是消息类型
// 要求:类型是结构体指针
func isMessage(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}
