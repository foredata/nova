package netx

import (
	"errors"
	"fmt"
	"io"

	"github.com/foredata/nova/pkg/bytex"
)

type CodecType = uint

// 常见已知CodecType枚举
const (
	CodecTypeUnknown  CodecType = 0
	CodecTypeText     CodecType = 1 // text/plain
	CodecTypeBinary   CodecType = 2 // application/octet-stream
	CodecTypeForm     CodecType = 3 // application/x-www-form-urlencoded
	CodecTypeJson     CodecType = 4
	CodecTypeXml      CodecType = 5
	CodecTypeProtobuf CodecType = 6
	CodecTypeThrift   CodecType = 7
	CodecTypeMsgpack  CodecType = 8
	CodecTypeAvro     CodecType = 9
	CodecTypeGob      CodecType = 10
)

// Codec 用于消息中body的编解码，常见的格式为Json,Xml,Protobuf,Thrift
//	不同协议间可能有不同的标识,比如http使用ContextType标识
type Codec interface {
	Type() CodecType
	Name() string
	Encode(p bytex.Buffer, msg interface{}) error
	Decode(p bytex.Buffer, msg interface{}) error
}

func init() {
	AddContentTypeMap(CodecTypeText, "text/plain")
	AddContentTypeMap(CodecTypeBinary, "application/octet-stream")
	AddContentTypeMap(CodecTypeForm, "application/x-www-form-urlencoded")
	AddContentTypeMap(CodecTypeJson, "application/json")
	AddContentTypeMap(CodecTypeXml, "application/xml")
	AddContentTypeMap(CodecTypeProtobuf, "application/protobuf")
	AddContentTypeMap(CodecTypeThrift, "application/thrift")
	AddContentTypeMap(CodecTypeMsgpack, "application/msgpack")
	AddContentTypeMap(CodecTypeMsgpack, "application/avro")
	AddContentTypeMap(CodecTypeMsgpack, "application/gob")
}

// http contentType <-> codecType
var (
	codecTypeToContentType = make(map[CodecType]string)
	contentTypeToCodecType = make(map[string]CodecType)
)

// AddContentTypeMap 添加MIME ContentType 到CodecType映射
func AddContentTypeMap(codecType CodecType, contentType string) {
	codecTypeToContentType[codecType] = contentType
	contentTypeToCodecType[contentType] = codecType
}

// GetContentType codecType转换为contentType
func GetContentType(ctype CodecType) string {
	return codecTypeToContentType[ctype]
}

// GetCodecType contentType转换为codecType
func GetCodecType(contentType string) CodecType {
	return contentTypeToCodecType[contentType]
}

var (
	// ErrNotFoundCodec not found codec err
	ErrNotFoundCodec = errors.New("not found codec")
)

var (
	gTypeMap = make(map[uint]Codec)
	gNameMap = make(map[string]Codec)
)

// RegisterCodec 添加Codec,非线程安全,通常仅在程序启动时注册
func RegisterCodec(c Codec) {
	if _, ok := gTypeMap[c.Type()]; ok {
		panic(fmt.Errorf("[codec] duplicate type=%+v", c.Type()))
	}
	if _, ok := gNameMap[c.Name()]; ok {
		panic(fmt.Errorf("[codec] duplicate  name=%+v", c.Name()))
	}
	gTypeMap[c.Type()] = c
	gNameMap[c.Name()] = c
}

// GetByType 通过类型获取Codec
func GetByType(t CodecType) Codec {
	if t != CodecTypeUnknown {
		return gTypeMap[t]
	}

	return nil
}

// GetByName 通过名字获取Codec
func GetByName(name string) Codec {
	if name != "" {
		return gNameMap[name]
	}

	return nil
}

// Encode 对消息进行编码
func Encode(ctype CodecType, msg interface{}) (bytex.Buffer, error) {
	enc := GetByType(ctype)
	if enc == nil {
		return nil, ErrNotFoundCodec
	}

	buf := bytex.NewBuffer()
	if err := enc.Encode(buf, msg); err != nil {
		buf.Clear()
		return nil, err
	}

	_, _ = buf.Seek(0, io.SeekStart)
	return buf, nil
}

// Decode 对消息进行解码
func Decode(buf bytex.Buffer, ctype CodecType, msg interface{}) error {
	dec := GetByType(ctype)
	if dec == nil {
		return ErrNotFoundCodec
	}

	return dec.Decode(buf, msg)
}
