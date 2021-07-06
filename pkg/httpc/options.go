package httpc

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

// 预定义常见类型
const (
	TypeJSON = "application/json"
	TypeXML  = "application/xml"
	TypeForm = "application/x-www-form-urlencoded"
	TypeHTML = "text/html"
	TypeText = "text/plain"
)

const (
	CharsetUTF8 = "UTF-8"
)

type CallOptions struct {
	BaseUrl     string            // 默认url地址
	ContentType string            // 编码格式,默认application/json
	Charset     string            // 编码格式,默认空,若非空则与contentType拼接到一起,比如 text/plain;charset=UTF-8
	Params      map[string]string // path中参数,例如: http://www.baidu.com/im/v1/chats/:chat_id
	Query       url.Values        // query参数,例如：http://www.baidu.com?aa=xxx&bb=xxx
	Header      http.Header       // 消息头
	Cookies     []*http.Cookie    //
	Retry       int               // 重试次数
	Hook        Hook              // 回调函数
}

var defaultCallOptions = &CallOptions{
	ContentType: TypeJSON,
}

func toCallOptions(opts ...*CallOptions) *CallOptions {
	if len(opts) > 0 {
		return opts[0]
	}

	return defaultCallOptions
}

func (o *CallOptions) AddParam(key string, value string) {
	if o.Params == nil {
		o.Params = make(map[string]string)
	}
	o.Params[key] = value
}

func (o *CallOptions) AddHeader(key string, value interface{}) {
	if o.Header == nil {
		o.Header = make(http.Header)
	}
	addValue(o.Header, key, value)
}

func (o *CallOptions) AddQuery(key string, value interface{}) {
	if o.Query == nil {
		o.Query = make(url.Values)
	}
	addValue(o.Query, key, value)
}

func (o *CallOptions) AddCookie(cookie *http.Cookie) {
	o.Cookies = append(o.Cookies, cookie)
}

func (o *CallOptions) AddAuthorization(auth string) {
	o.AddHeader("Authorization", auth)
}

func (o *CallOptions) AddBasicAuth(username, password string) {
	// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
	// "To receive authorization, the client sends the userid and password,
	// separated by a single colon (":") character, within a base64
	// encoded string in the credentials."
	// It is not meant to be urlencoded.
	auth := username + ":" + password
	baseAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	o.AddHeader("Authorization", "Basic "+baseAuth)
}

func (o *CallOptions) AddBearAuth(auth string) {
	o.AddHeader("Authorization", "Bearer "+auth)
}

func (o *CallOptions) AddXJwtToken(token string) {
	o.AddHeader("X-Jwt-Token", token)
}

func (o *CallOptions) AddXAuthToken(token string) {
	o.AddHeader("X-Auth-Token", token)
}

// url.Values, http.Header
func addValue(dict map[string][]string, key string, value interface{}) {
	switch v := value.(type) {
	case string:
		dict[key] = append(dict[key], v)
	case []string:
		dict[key] = append(dict[key], v...)
	default:
		dict[key] = append(dict[key], fmt.Sprintf("%+v", v))
	}
}
