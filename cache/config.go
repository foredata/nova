package cache

import (
	"time"

	"github.com/foredata/nova/encoding"
)

const (
	defaultMaxSize   = 5000
	defaultJitterTTL = time.Minute * 5
)

// CreatorFunc 创建Entity,用于Get时，调用Unmarshal entity
type CreatorFunc func() interface{}

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// KeyModifier 修正key,添加前缀或者后缀
type KeyModifier func(key string) string

// Config cache 配置信息
type Config struct {
	Modifier  KeyModifier            // 修正key,添加前缀或后缀,默认不修改key
	Filter    Filter                 // 预先忽略不存在的key,防止击穿,可使用BloomFilter
	AbsentTTL time.Duration          // 用于标识是否缓存不存在的key,若为0则不缓存无效的key,默认为0,避免击穿
	JitterTTL time.Duration          // 当写入时,如果带有ttl,则会自动增加一个抖动时间,避免雪崩
	MaxSize   int                    // 限制最大条数
	Codec     encoding.Codec         // 编解码,默认使用json
	Creator   CreatorFunc            // 创建entity回调
	OnEvict   EvictCallback          // 驱逐时回调
	Extra     map[string]interface{} // 扩展配置
}

// Get 通过key获取extra配置
func (c *Config) Get(key string) interface{} {
	return c.Extra[key]
}

// Prefix key添加前缀
func Prefix(prefix string) KeyModifier {
	return func(key string) string {
		return prefix + key
	}
}

// Suffix key添加后缀
func Suffix(suffix string) KeyModifier {
	return func(key string) string {
		return key + suffix
	}
}

// defaultKeyModifier 默认不做任何处理
func defaultKeyModifier(key string) string {
	return key
}
