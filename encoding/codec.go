package encoding

import (
	"encoding/json"
)

// Marshaler 通用Marshal接口
type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

// Unmarshaler 通用Unmarshal接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// Codec 编解码,默认使用json
type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

func NewJsonCodec() Codec {
	return &jsonCodec{}
}

type jsonCodec struct {
}

func (c *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
