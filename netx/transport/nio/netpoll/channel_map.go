package netpoll

import "sync"

type channelMap struct {
	mux  sync.RWMutex
	dict map[FD]Channel
}

func (cm *channelMap) Init() {
	cm.dict = make(map[FD]Channel)
}

func (cm *channelMap) Get(fd FD) Channel {
	cm.mux.RLock()
	defer cm.mux.RUnlock()
	return cm.dict[fd]
}

func (cm *channelMap) Add(ch Channel) {
	cm.mux.Lock()
	cm.dict[ch.Fd()] = ch
	cm.mux.Unlock()
}

func (cm *channelMap) Del(fd FD) {
	cm.mux.Lock()
	delete(cm.dict, fd)
	cm.mux.Unlock()
}
