package netx

import "net/http"

const (
	StatusInternalServerError = http.StatusInternalServerError
	StatusTimeout             = http.StatusRequestTimeout
)

const (
	MethodUnknown Method = iota
	MethodAny
	MethodGet
	MethodHead
	MethodPost
	MethodPut
	MethodPatch
	MethodDelete
	MethodConnect
	MethodOptions
	MethodTrace
)

// Method http method
type Method uint8

func (m Method) IsValid() bool {
	return m >= MethodGet && m <= MethodTrace
}

func (m Method) String() string {
	switch m {
	case MethodAny:
		return "Any"
	case MethodGet:
		return "GET"
	case MethodHead:
		return "HEAD"
	case MethodPost:
		return "POST"
	case MethodPut:
		return "PUT"
	case MethodPatch:
		return "PATCH"
	case MethodDelete:
		return "DELETE"
	case MethodConnect:
		return "CONNECT"
	case MethodOptions:
		return "OPTIONS"
	case MethodTrace:
		return "TRACE"
	default:
		return "Unknown"
	}
}

func ParseMethod(str string) Method {
	switch str {
	case "GET":
		return MethodGet
	case "HEAD":
		return MethodHead
	case "POST":
		return MethodPost
	case "PUT":
		return MethodPut
	case "PATCH":
		return MethodPatch
	case "DELETE":
		return MethodDelete
	case "CONNECT":
		return MethodConnect
	case "OPTIONS":
		return MethodOptions
	case "TRACE":
		return MethodTrace
	default:
		return MethodUnknown
	}
}

// Params path中参数,values长度可能比keys大,但不会小,用于cache value
type Params struct {
	keys   []string
	values []string
}

func (p *Params) Len() int {
	return len(p.keys)
}

// Get 通过名字获取数据
func (p *Params) Get(key string) string {
	for i := 0; i < len(p.keys); i++ {
		if p.keys[i] == key {
			return p.values[i]
		}
	}

	return ""
}

// Keys 返回所有key
func (p *Params) Keys() []string {
	return p.keys
}

// Values 返回所有value
func (p *Params) Values() []string {
	return p.values
}

// Reset 重置数据
func (p *Params) Reset(keys, values []string) {
	p.keys = keys
	p.values = values
}
