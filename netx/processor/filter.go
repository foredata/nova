package processor

import (
	"io"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

// NewFilter .
func NewFilter(executor netx.Executor, provider Provider, detector netx.Detector) netx.Filter {
	return &filter{
		processor: New(executor, provider),
		detector:  detector,
	}
}

type filter struct {
	netx.BaseFilter
	processor netx.Processor // 处理回调
	detector  netx.Detector  // 协议探测
}

func (f *filter) Name() string {
	return "processor"
}

func (f *filter) HandleRead(ctx netx.FilterCtx) error {
	conn := ctx.Conn()
	data, ok := ctx.Data().(bytex.Buffer)
	if !ok {
		return nil
	}

	// detect protocol
	proto, _ := conn.Protocol().(netx.Protocol)
	if proto == nil {
		proto = f.detector.Detect(data)
		if proto == nil {
			return nil
		}
		conn.SetProtocol(proto)
	}

	// decode protocol
	frame, err := proto.Decode(conn, data)
	if err != nil || frame == nil {
		return err
	}

	// 丢弃已经解析过的数据
	data.Discard()

	// process message
	return f.processor.Process(conn, frame)
}

func (f *filter) HandleWrite(ctx netx.FilterCtx) error {
	data := ctx.Data()
	if data == nil {
		return nil
	}

	conn := ctx.Conn()
	proto, _ := conn.Protocol().(netx.Protocol)
	if proto == nil {
		proto = f.detector.Default()
		conn.SetProtocol(proto)
	}

	switch x := data.(type) {
	case netx.Packet:
		wt, err := encodePacket(proto, conn, x)
		if err != nil {
			return err
		}
		ctx.SetData(wt)
		return nil
	case netx.Frame:
		buff, err := proto.Encode(conn, x)
		if err != nil {
			return err
		}
		_, _ = buff.Seek(0, io.SeekStart)
		ctx.SetData(buff)
		return nil
	case bytex.Buffer:
		_, _ = x.Seek(0, io.SeekStart)
	}

	return nil
}

func encodePacket(proto netx.Protocol, conn netx.Conn, packet netx.Packet) (netx.WriterTo, error) {
	body := packet.Body()
	var payload bytex.Buffer
	var chunked bool
	if body != nil {
		payload, _ = body.ReadFast(false)
		chunked = !body.End()
	}

	if !chunked {
		// 不需要分块传输
		header := packet.Header()
		if header == nil {
			header = netx.NewHeader()
		}
		header.Merge(packet.Trailer())

		if payload != nil {
			_, _ = payload.Seek(0, io.SeekStart)
		}

		frame := netx.NewFrame(netx.FrameTypeHeader, true, 0, packet.Identifier(), header, payload)
		buf, err := proto.Encode(conn, frame)
		if buf != nil {
			_, _ = buf.Seek(0, io.SeekStart)
		}
		return buf, err
	}

	// TODO: chunked encode

	return nil, netx.ErrNotSupport
}
