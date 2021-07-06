package local

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"
	"time"

	"github.com/foredata/nova/netx/registry"
)

func newWatcher(rootDir string) *localWatcher {
	w := &localWatcher{rootDir: rootDir}
	return w
}

type watchFile struct {
	nodeId  string            //
	hash    string            //
	modtime time.Time         // 最近更新时间
	service *registry.Service //
}

// localWatcher 定时轮训的方式diff文件变化
type localWatcher struct {
	rootDir string
	files   map[string]*watchFile
	eventCh chan *registry.Event
	ticker  *time.Ticker
}

func (w *localWatcher) Next() (*registry.Event, error) {
	ev := <-w.eventCh
	return ev, nil
}

func (w *localWatcher) Start() {
	w.files = make(map[string]*watchFile)
	w.eventCh = make(chan *registry.Event)
	w.ticker = time.NewTicker(time.Second)
	go func() {
		for ; true; <-w.ticker.C {
			w.Diff()
		}
	}()
}

func (w *localWatcher) Stop() {
	if w.ticker == nil {
		return
	}
	w.ticker.Stop()
	close(w.eventCh)
	w.ticker = nil
	w.eventCh = nil
}

func (w *localWatcher) Diff() {
	files, err := ioutil.ReadDir(w.rootDir)
	if err != nil {
		return
	}

	newFiles := make(map[string]*watchFile, len(w.files))
	now := time.Now()
	for _, info := range files {
		if info.IsDir() || !isServiceFile(info.Name()) {
			continue
		}

		nodeId := toNodeId(info.Name())
		wf := w.files[info.Name()]

		filename := w.rootDir + info.Name()
		ent, data, err := fromFile(filename)
		if err != nil || now.After(ent.Expired) {
			_ = os.Remove(filename)
			if wf != nil {
				w.eventCh <- &registry.Event{Id: nodeId, Type: registry.EventDelete, Service: wf.service}
				delete(w.files, info.Name())
			}
			continue
		}

		if wf == nil {
			// 新服务
			wf = &watchFile{nodeId: nodeId, hash: calcHash(data), modtime: info.ModTime(), service: ent.Service}
			newFiles[info.Name()] = wf
			w.eventCh <- &registry.Event{Id: nodeId, Type: registry.EventCreate, Service: ent.Service}
			continue
		}

		// diff change
		hash := calcHash(data)
		if wf.modtime != info.ModTime() || hash != wf.hash {
			wf.hash = hash
			wf.modtime = info.ModTime()
			wf.service = ent.Service
			w.eventCh <- &registry.Event{Id: nodeId, Type: registry.EventUpdate, Service: ent.Service}
		}

		newFiles[info.Name()] = wf
		delete(w.files, info.Name())
	}

	// check deleted
	for key, f := range w.files {
		w.eventCh <- &registry.Event{Id: f.nodeId, Type: registry.EventDelete, Service: f.service}
		delete(w.files, key)
	}

	w.files = newFiles
}

func calcHash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
