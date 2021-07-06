package netpoll

// New 新建poll
func New() (Poller, error) {
	return newPoller()
}

const maxEventNum = 1024

// Event 事件mask枚举,类似Epoll中定义
type Event uint

const (
	EventIn = 0x01 // 可读
	//EventPri   = 0x02 // 紧急事件
	EventOut = 0x04 // 可写
	EventErr = 0x08 // 出错
	//EventHup   = 0x10 // 关闭
	EventInOut = EventIn | EventOut
)

// Is 判断是否含有某个状态
func (events Event) Is(target Event) bool {
	return events&target != 0
}

// Channel 代表socket
type Channel interface {
	Fd() FD
	OnEvent(events Event)
}

// Poller epoll/kqueue/select
// kqueue&epoll控制事件的注意点: https://www.cnblogs.com/herm/archive/2010/11/12/2773963.html
// 深入学习理解 IO 多路复用: https://segmentfault.com/a/1190000022352273
type Poller interface {
	Close() error                               // 关闭
	Wakeup() error                              // 唤醒wait
	Wait() error                                // 等待触发事件,每次只会触发一次,返回io.EOF代表结束
	Insert(channel Channel, events Event) error // 添加事件
	Modify(channel Channel, events Event) error // 更新事件
	Delete(channel Channel) error               // 删除事件
}
