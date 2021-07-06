package http1

import (
	"io"
	"strings"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
	"github.com/foredata/nova/pkg/unique"
)

var gHttpProcotol = &http1Protocol{}

// New .
func New() netx.Protocol {
	return gHttpProcotol
}

var (
	// kConnKeyHttpDecoder conn中unique key
	kConnKeyHttpDecoder = unique.NewKey(netx.KeyGroupConn, "http1-decoder")
)

// 实现protocol.x,h1不支持多路复用,解析有状态
// https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol
// https://en.wikipedia.org/wiki/Chunked_transfer_encoding
// header多行(https://stackoverflow.com/questions/31237198/is-it-possible-to-include-multiple-crlfs-in-a-http-header-field)
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Transfer-Encoding
// https://blog.csdn.net/liuxiao723846/article/details/107433395
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages
// https://www.kancloud.cn/spirit-ling/http-study/636182
// https://www.cnblogs.com/lovelacelee/p/5385683.html
// https://www.cnblogs.com/arnoldlu/p/6497837.html
type http1Protocol struct {
}

func (*http1Protocol) Name() string {
	return "http1"
}

// Detect http协议以行解析,并不容易探测,需要读取完整第一行才能获取
// 目前通过判断GET,POST等识别
// GET /index.html HTTP/1.1
// [协议升级机制] https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Protocol_upgrade_mechanism
func (*http1Protocol) Detect(p bytex.Peeker) bool {
	const bufSize = 10
	var buf [bufSize]byte
	n, err := p.Peek(buf[:])
	if err != nil || n != bufSize {
		return false
	}

	idx := strings.IndexByte(string(buf[:]), ' ')
	if idx == -1 {
		return false
	}
	text := string(buf[:idx])
	method := netx.ParseMethod(text)
	return method != netx.MethodUnknown
}

// Decode 解析http协议,有状态,每个Conn一个decoder
func (*http1Protocol) Decode(conn netx.Conn, buf bytex.Buffer) (netx.Frame, error) {
	dec := conn.Attributes().Get(kConnKeyHttpDecoder, func() interface{} {
		return newDecoder()
	}).(*decoder)

	frame, err := dec.Decode(buf)
	if err == io.EOF {
		err = nil
	}

	return frame, err
}

func (*http1Protocol) Encode(conn netx.Conn, frame netx.Frame) (bytex.Buffer, error) {
	enc := encoder{}
	return enc.Encode(frame)
}
