package codec

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/codec/gob"
	"github.com/foredata/nova/netx/codec/json"
	"github.com/foredata/nova/netx/codec/protobuf"
	"github.com/foredata/nova/netx/codec/xml"
)

// init 需要在启动时手动注册一下,server/client默认已经注册过
func init() {
	Register(json.New())
	Register(xml.New())
	Register(protobuf.New())
	Register(gob.New())
}

// Register 添加Codec,非线程安全,通常仅在程序启动时注册
func Register(c netx.Codec) {
	netx.RegisterCodec(c)
}
