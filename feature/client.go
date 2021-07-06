package feature

import "context"

const (
	ToggleOn  = "on"
	ToggleOff = "off"
)

// Attributes 业务自定义数据,可根据属性判断是否开启对应feature
type Attributes map[string]interface{}

// Client 抽象接口,实现可以基于本地配置，远程配置，Database，或第三方平台SDK
type Client interface {
	// IsToggled 类似Evaluate，但返回结果只有on/off两个分支
	IsToggled(ctx context.Context, key string, attributes Attributes, defaultVal bool) (bool, error)
	// Evaluate 根据用户的attributes计算对应的key所命中的分支，需要提供一个默认值,当返回错误时，则返回默认值
	Evaluate(ctx context.Context, key string, attributes Attributes, defaultVal string) (string, error)
}
