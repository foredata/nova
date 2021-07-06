package discovery

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"sort"
)

// Resolver 用于解析service
//	可以是基于registry的服务发现
//	也可以是基于DNS的域名解析
//	或者返回固定IP
type Resolver interface {
	Name() string
	Resolve(ctx context.Context, service string) (*Result, error)
	Close() error
}

func NewResult(instances []Instance) *Result {
	r := &Result{}
	r.SetInstances(instances)
	return r
}

type Result struct {
	instances []Instance
	hashCode  uint32
}

func (r *Result) Empty() bool {
	return len(r.instances) == 0
}

func (r *Result) Len() int {
	return len(r.instances)
}

func (r *Result) Instances() []Instance {
	return r.instances
}

func (r *Result) SetInstances(ins []Instance) {
	r.instances = ins
	r.hashCode = calcHashCode(ins)
}

func (r *Result) Remove(id string) {
	for idx, ins := range r.instances {
		if ins.Id() == id {
			// order is not important
			last := len(r.instances) - 1
			r.instances[last], r.instances[idx] = r.instances[idx], r.instances[last]
			r.instances = r.instances[:last]
			r.hashCode = calcHashCode(r.instances)
			break
		}
	}
}

// Upsert 添加或更新
func (r *Result) Upsert(ins Instance) {
	for idx, old := range r.instances {
		if old.Id() == ins.Id() {
			r.instances[idx] = ins
			r.hashCode = calcHashCode(r.instances)
			return
		}
	}

	r.instances = append(r.instances, ins)
	r.hashCode = calcHashCode(r.instances)
}

func (r *Result) HashCode() uint32 {
	return r.hashCode
}

func calcHashCode(instances []Instance) uint32 {
	if len(instances) == 0 {
		return 0
	}

	h := fnv.New32()

	for _, ins := range instances {
		h.Write([]byte(ins.Addr()))
		if ins.Weight() != 0 {
			h.Write(uint32ToBytes(ins.Weight()))
		}
		tags := mapToOrderedSlice(ins.Tags())
		if len(tags) > 0 {
			for _, t := range tags {
				h.Write([]byte(t))
			}
		}
	}

	return h.Sum32()
}

func uint32ToBytes(v uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(v))
	return bs
}

func mapToOrderedSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	result := make([]string, 0, len(m)*2)
	for k := range m {
		if k == "" {
			continue
		}
		result = append(result, k)
	}
	sort.Strings(result)
	for _, k := range result {
		v := m[k]
		if v == "" {
			continue
		}
		result = append(result, v)
	}
	return result
}

type Instance interface {
	Id() string
	Addr() string
	Weight() uint32
	Tags() map[string]string
}

func NewInstance(id, addr string, weight uint32, tags map[string]string) Instance {
	if id == "" {
		id = addr
	}
	return &instance{id: id, addr: addr, weight: weight, tags: tags}
}

type instance struct {
	id     string
	addr   string
	weight uint32
	tags   map[string]string
}

func (ins *instance) Id() string {
	return ins.id
}

func (ins *instance) Addr() string {
	return ins.addr
}

func (ins *instance) Weight() uint32 {
	return ins.weight
}

func (ins *instance) Tags() map[string]string {
	return ins.tags
}
