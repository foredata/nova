package bytex

import (
	"math"
	"sync"
)

// 64byte-1MB
var gPool = NewPool(6, 20)

// SetPool 设置默认的Pool,非线程安全,仅可以在程序启动前设置
func SetPool(p *Pool) {
	gPool = p
}

// Alloc 申请内存
func Alloc(size int) []byte {
	return gPool.Get(size)
}

// Free 释放内存
func Free(data []byte) bool {
	return gPool.Put(data)
}

// NewPool 创建Pool
func NewPool(minPow, maxPow int) *Pool {
	p := &Pool{}
	p.init(minPow, maxPow)
	return p
}

// Pool 内存池
// https://studygolang.com/articles/16282
// https://stackoverflow.com/questions/53613556/is-go-sync-pool-much-slower-than-make
// https://github.com/golang/go/issues/16323
type Pool struct {
	pools   []sync.Pool
	minPow  int
	minSize int
	maxSize int
}

func (p *Pool) init(minPow, maxPow int) {
	p.pools = make([]sync.Pool, maxPow-minPow+1)
	for i := minPow; i <= maxPow; i++ {
		idx := i - minPow
		size := 1 << i
		p.pools[idx] = sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}

	p.minPow = minPow
	p.minSize = 1 << minPow
	p.maxSize = 1 << maxPow
}

// Get 获取byte
func (p *Pool) Get(size int) []byte {
	if size < p.minSize || size > p.maxSize {
		b := make([]byte, size)
		return b
	}

	nextSize := roundUpPowerOfTwo(size)

	power := int(math.Log2(float64(nextSize)))
	index := power - p.minPow

	return p.pools[index].Get().([]byte)
}

// Put 释放byte
func (p *Pool) Put(data []byte) bool {
	size := len(data)
	if size < p.minSize || size > p.maxSize {
		return false
	}

	if !isPowerOfTwo(size) {
		return false
	}

	power := int(math.Log2(float64(size)))
	index := power - p.minPow

	// nolint:staticcheck
	p.pools[index].Put(data)
	return true
}

func isPowerOfTwo(v int) bool {
	return v > 0 && v&(v-1) == 0
}

func roundUpPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}
