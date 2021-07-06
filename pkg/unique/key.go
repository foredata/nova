package unique

import (
	"fmt"
	"sync"
)

var (
	gKeyGroup = make(map[string]int) // key group当前最大ID
	gKeyMap   = make(map[string]Key) // 记录所有key对应ID
	gKeyMux   = sync.Mutex{}         // 全局锁
)

// NewKey 通过group,name创建key
func NewKey(group, name string) Key {
	gKeyMux.Lock()
	defer gKeyMux.Unlock()
	k := fmt.Sprintf("%s_%s", group, name)
	if res, ok := gKeyMap[k]; ok {
		return res
	}

	id := gKeyGroup[group]
	res := &key{id: id, name: name, group: group}
	gKeyGroup[group] = id + 1
	gKeyMap[k] = res
	return res
}

// Key 唯一key,将string映射到唯一索引
// 相同group下，id从零自增，可以作为数组下标所为索引,不同机器下不能保证id相同,故不能用id序列化
// 通常为全局静态变量
type Key interface {
	ID() int
	Name() string
	Group() string
	String() string
}

type key struct {
	id    int
	name  string
	group string
}

func (k *key) ID() int {
	return k.id
}

func (k *key) Name() string {
	return k.name
}

func (k *key) Group() string {
	return k.group
}

func (k *key) String() string {
	return fmt.Sprintf("%s_%s", k.group, k.name)
}
