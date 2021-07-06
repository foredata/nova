package local

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/foredata/nova/netx/registry"
)

func New(opts ...registry.Option) registry.Registry {
	l := &localRegistry{}
	if err := l.Init(); err != nil {
		return nil
	}
	return l
}

const (
	// 根目录
	rootDir = ".karas/services"
)

var (
	errInvalidRootDir = errors.New("invalid root dir")
)

// 基于本地文件的服务注册发现
type localRegistry struct {
	rootDir string
}

func (l *localRegistry) Init() error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := path.Join(dirname, rootDir)
	if err := os.MkdirAll(l.rootDir, os.ModePerm); err != nil {
		return err
	}
	l.rootDir = dir
	return nil
}

func (l *localRegistry) Name() string {
	return "local"
}

func (l *localRegistry) Register(ctx context.Context, service *registry.Service, ttl time.Duration) error {
	if l.rootDir == "" {
		return errInvalidRootDir
	}

	if len(service.Nodes) == 0 {
		return fmt.Errorf("invalid service node num")
	}

	node := service.Nodes[0]
	filename := toFilename(l.rootDir, node.ID)

	expired := time.Now().Add(ttl)
	ent := &entry{Expired: expired, Service: service}

	data, err := ent.Encode()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, os.ModePerm)
}

func (l *localRegistry) Deregister(ctx context.Context, service *registry.Service) error {
	if l.rootDir == "" {
		return errInvalidRootDir
	}
	node := service.Nodes[0]
	filename := toFilename(l.rootDir, node.ID)
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("deregister fail, %+v", err)
	}

	return nil
}

func (l *localRegistry) Get(ctx context.Context, service string) ([]*registry.Service, error) {
	files, err := ioutil.ReadDir(l.rootDir)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make([]*registry.Service, 0, len(files))
	for _, info := range files {
		if info.IsDir() || !isServiceFile(info.Name()) {
			continue
		}
		filename := l.rootDir + info.Name()
		ent, _, err := fromFile(filename)
		if err != nil {
			_ = os.Remove(filename)
			continue
		}

		if now.After(ent.Expired) {
			_ = os.Remove(filename)
			continue
		}

		if ent.Service.Name != service {
			continue
		}
		result = append(result, ent.Service)
	}

	return result, nil
}

func (l *localRegistry) List(ctx context.Context) ([]*registry.Service, error) {
	files, err := ioutil.ReadDir(l.rootDir)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make([]*registry.Service, 0, len(files))
	for _, info := range files {
		if info.IsDir() || !isServiceFile(info.Name()) {
			continue
		}
		filename := l.rootDir + info.Name()
		ent, _, err := fromFile(filename)
		if err != nil {
			_ = os.Remove(filename)
			continue
		}

		if now.After(ent.Expired) {
			_ = os.Remove(filename)
			continue
		}

		result = append(result, ent.Service)
	}

	return result, nil
}

func (l *localRegistry) Watch(ctx context.Context) (registry.Watcher, error) {
	w := newWatcher(l.rootDir)
	return w, nil
}

func (l *localRegistry) Close() error {
	return nil
}
