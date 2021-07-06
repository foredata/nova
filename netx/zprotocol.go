package netx

import (
	"context"
	"net/url"

	"github.com/foredata/nova/netx/body"
	"github.com/foredata/nova/netx/metadata"
	"github.com/foredata/nova/pkg/bytex"
)

// 常见Header
const (
	XLogId   = "X-Log-Id"
	XTraceId = "X-Trace-Id"
)

type Header = metadata.Metadata

func NewHeader() Header {
	return metadata.New()
}

func NewIdentifier() *Identifier {
	return &Identifier{}
}

// MsgType 消息类型,定义同thrift
// https://github.com/apache/thrift/blob/master/doc/specs/thrift-binary-protocol.md
type MsgType uint8

const (
	MsgTypeCall      MsgType = 1 // Request 请求消息
	MsgTypeReply     MsgType = 2 // Response 应答消息
	MsgTypeException MsgType = 3 // 异常消息, StatusCode != 0 && StatusCode != 200
	MsgTypeOneway    MsgType = 4 // Request 无效应答
)

// Identifier 消息标识,只会在header中使用
type Identifier struct {
	Version    uint     // 版本信息,对于http1则对应[09,10,11]
	IsResponse bool     // 是否是应答消息
	IsOneway   bool     // 是否需要应答
	SeqID      uint32   // 动态唯一ID,用于查询Response回调
	CmdID      uint32   // CmdID,用于查询Request回调
	Method     Method   // http method
	Service    string   // http host
	URI        string   // http URI
	Codec      uint32   // payload编码,区别于Content-Type,Codec只支持有限的编码方式
	StatusCode int32    // response status code
	StatusInfo string   // response status info
	Params     Params   // 从path中解析获得的参数
	url        *url.URL // 解析uri获得
}

func (ident *Identifier) URL() *url.URL {
	if ident.url != nil {
		return ident.url
	}
	if ident.URI != "" {
		u, err := url.Parse(ident.URI)
		if err == nil {
			ident.url = u
		}
	}

	return ident.url
}

func (ident *Identifier) MsgType() MsgType {
	if ident.IsResponse {
		if ident.StatusCode != 0 && ident.StatusCode != 200 {
			return MsgTypeException
		}
		return MsgTypeReply
	} else {
		if ident.IsOneway {
			return MsgTypeOneway
		}
		return MsgTypeCall
	}
}

type FrameType uint8

const (
	FrameTypeHeader  = 0 // 消息头
	FrameTypeData    = 1
	FrameTypeTrailer = 2
)

// Frame 最底层消息帧,一个消息可以由一帧组成,也可以由多帧组成
//	一个消息通常由三部分组成,header,body,trailer
//	EndFlag标记是否是消息的最后一帧
//	StreamID,用于组装成Message
//	Header帧会包含Identifier,Header,Payload
//	Data帧只会使用Payload
//	Trailer帧只会使用Trailer
//
//	Header和Trailer只能有一个,Data帧可以有多个
// 参考http2:
// https://hpbn.co/http2/
// https://halfrost.com/http2-http-semantics/
// https://www.cnblogs.com/yudar/p/4642603.html
type Frame interface {
	Recycler
	Type() FrameType
	EndFlag() bool
	StreamID() uint32
	SetStreamID(id uint32)
	Identifier() *Identifier
	SetIdentifier(v *Identifier)
	Header() Header
	SetHeader(v Header)
	Trailer() Header
	SetTrailer(v Header)
	Payload() bytex.Buffer
	SetPayload(v bytex.Buffer)
}

// Body see body.Body
type Body = body.Body

// BodyWriter see body.Writer
type BodyWriter = body.Writer

// Packet 完整消息,由多个Frame组合而成,
// 	同Request和Response底层使用相同结构,不同的是Packet通常由底层系统使用,而另外两个对用户使用更方便
//	在通信模型上通常分为两种:
//	1: 消息模式:内容为特定编码的消息结构,用于RPC通信
//	2: Chunk模式:内容通常为文件内容,用于收发文件等
type Packet interface {
	Recycler
	Identifier() *Identifier
	SetIdentifier(v *Identifier)
	Header() Header
	SetHeader(v Header)
	Trailer() Header
	SetTrailer(v Header)
	Body() Body
	SetBody(v Body)
}

// Request see http.Request
type Request interface {
	Recycler
	Version() uint
	SetVersion(v uint)
	SeqID() uint32
	SetSeqID(v uint32)
	Codec() uint32
	SetCodec(v uint32)
	IsOneway() bool
	SetOneway(v bool)
	CmdID() uint32
	SetCmdID(v uint32)
	Service() string
	SetService(v string)
	URI() string
	SetURI(v string)
	URL() *url.URL
	Params() Params
	Method() Method
	SetMethod(v Method)
	Header() Header
	SetHeader(v Header)
	Trailer() Header
	SetTrailer(v Header)
	Body() Body
	SetBody(v Body)
	Encode(codecType CodecType, msg interface{}) error
	Decode(msg interface{}) error
}

// Response see http.Response
type Response interface {
	Recycler
	Version() uint
	SetVersion(v uint)
	SeqID() uint32
	SetSeqID(v uint32)
	Codec() uint32
	SetCodec(v uint32)
	StatusCode() int32
	StatusInfo() string
	SetStatus(code int32, info string)
	Header() Header
	SetHeader(v Header)
	Trailer() Header
	SetTrailer(v Header)
	Body() Body
	SetBody(v Body)
	Encode(codecType CodecType, msg interface{}) error
	Decode(msg interface{}) error
}

// Protocol 通信协议编解码
//	类似HTTP2协议,这里区分了Frame和Message两个概念,但弱化了http2协议
//	Protocol只负责Frame的编解码,并不负责Frame到Message的组包和拆包
// multiplexing
// https://github.com/apache/incubator-brpc/blob/master/docs/cn/baidu_std.md
type Protocol interface {
	Name() string
	Detect(data bytex.Peeker) bool
	Decode(conn Conn, data bytex.Buffer) (Frame, error)
	Encode(conn Conn, frame Frame) (bytex.Buffer, error)
}

// Detector 用于自动探测协议,某些协议有magic number,可以方便的感知协议类型,某些则不支持
//	服务端需要探测协议,但仅需要探测一次即可,便于自动识别http,dubbo,grpc等协议
//	客户端则不需要探测协议,因为调用方是知道使用哪种协议
type Detector interface {
	Detect(bytex.Peeker) Protocol // 服务端探测协议
	Default() Protocol            // 默认协议,客户端不需要
}

// Processor 用于接收Frame然后拼装成Packet,并调用回调函数
// 通常有以下几种模式:
//	1:普通消息，一次能收发完完整消息,处理比较简单,没有并发问题,比如ping-pong，oneway模式rpc
//	2:流式消息，一个包由多帧组成,有并发问题，需要保证同时只能有一个线程在处理,但可能会被多次触发
//	常见场景：a:文件分片传输,b:stream rpc
type Processor interface {
	Process(conn Conn, frame Frame) error
}

// Callback processor回调函数
//	对于服务端packet为Request
//	对于客户端packet为Response
type Callback func(conn Conn, packet Packet) error

// Endpoint represent one method for calling from remote.
type Endpoint func(ctx context.Context, req Request) (Response, error)

// Middleware deal with input Endpoint and output Endpoint.
type Middleware func(Endpoint) Endpoint

// Chain connect middlewares into one middleware.
func Chain(mws []Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Apply 将Endpoint添加middleware,并返回最终的Endpoint
func Apply(endpoint Endpoint, mws []Middleware) Endpoint {
	for i := len(mws) - 1; i >= 0; i-- {
		endpoint = mws[i](endpoint)
	}

	return endpoint
}
