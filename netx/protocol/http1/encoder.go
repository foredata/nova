package http1

import (
	"strconv"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

// https://cloud.google.com/apigee/docs/api-platform/antipatterns/multi-value-http-headers
// https://stackoverflow.com/questions/3096888/standard-for-adding-multiple-values-of-a-single-http-header-to-a-request-or-resp/38406581
type encoder struct {
}

func (e *encoder) Encode(frame netx.Frame) (bytex.Buffer, error) {
	switch frame.Type() {
	case netx.FrameTypeHeader:
		return e.writeHeader(frame)
	case netx.FrameTypeTrailer:
		return e.writeTrailer(frame)
	case netx.FrameTypeData:
		return e.writeData(frame)
	default:
		return nil, netx.ErrNotSupport
	}
}

func (e *encoder) writeHeader(frame netx.Frame) (bytex.Buffer, error) {
	buf := bytex.NewBuffer()
	ident := frame.Identifier()
	if ident == nil {
		return nil, netx.ErrInvalidFrame
	}
	version := toHttpVersion(ident.Version)
	header := frame.Header()
	payload := frame.Payload()

	// 注:严格匹配Header
	if header.Get(HeaderContentType) == "" && ident.Codec != 0 {
		contentType := netx.GetContentType(netx.CodecType(ident.Codec))
		header.Set(HeaderContentType, contentType)
	}

	if !frame.EndFlag() {
		header.Set(HeaderTransferEncoding, transferEncodingChunked)
	} else {
		contentLen := 0
		if payload != nil && !payload.Empty() {
			contentLen = payload.Len()
		}
		header.Set(HeaderContentLength, strconv.Itoa(contentLen))
	}

	// identifier
	var err error
	if ident.IsResponse {
		code, info := toHttpStatus(ident)
		err = bytex.Writef(buf, "%s %d %s\r\n", version, code, info)
	} else {
		err = bytex.Writef(buf, "%s %s %s\r\n", ident.Method.String(), ident.URI, version)
	}

	if err != nil {
		return nil, err
	}

	// header
	writeHeader(buf, header)

	_ = bytex.Write(buf, kCRLF)

	// write payload
	if payload != nil {
		_ = buf.Append(payload)
	}

	return buf, nil
}

func (e *encoder) writeTrailer(frame netx.Frame) (bytex.Buffer, error) {
	buf := bytex.NewBuffer()
	trailer := frame.Trailer()
	if len(trailer) == 0 {
		return nil, nil
	}

	writeHeader(buf, trailer)
	_ = bytex.Write(buf, kCRLF)

	return buf, nil
}

func (e *encoder) writeData(frame netx.Frame) (bytex.Buffer, error) {
	payload := frame.Payload()

	buf := bytex.NewBuffer()
	if payload != nil && payload.Len() > 0 {
		_ = bytex.Writef(buf, "")
		_ = buf.Append(payload)
	}
	_ = bytex.Write(buf, kCRLF)

	return buf, nil
}

func writeHeader(buf bytex.Buffer, header netx.Header) {
	header.Walk(func(key string, values []string) bool {
		for _, v := range values {
			_ = bytex.Writef(buf, "%s: %s\r\n", key, v)
		}
		return true
	})
}
