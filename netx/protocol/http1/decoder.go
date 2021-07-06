package http1

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

var (
	errInvalidHttpIdent   = errors.New("invalid http ident")
	errInvalidHttpMethod  = errors.New("invalid http method")
	errInvalidHttpVersion = errors.New("invalid http version")
	errInvaidHttpHeader   = errors.New("invalid http header")
	errLineTooLong        = errors.New("header line too long")
	errNoContentLength    = errors.New("no content length")
)

const (
	headerSize    = 8
	maxLineLength = 4096 // assumed <= bufio.defaultBufSize
)

type state uint8

const (
	stateIdent   state = iota // 读取首行标识
	stateHeader               // 读取header
	stateBody                 // 读取普通消息体
	stateChunk                // 读取分块数据
	stateTrailer              // 读取trailer
)

func newDecoder() *decoder {
	dec := &decoder{}
	return dec
}

// decoder http解析器
// 由于http不能通过长度判断消息包是否完整,需要缓存当前解析状态
// 参考:
//	http.ReadRequest
//	http.ReadResponse
//	ReadMIMEHeader
//	https://www.kancloud.cn/spirit-ling/http-study/636182
// 	https://paper.seebug.org/1048/
type decoder struct {
	state    state            // 当前解析状态
	ident    *netx.Identifier // 当前解析的identify
	header   netx.Header      // 当前消息头
	trailer  netx.Header      // 当前trailer
	line     bytex.Buffer     // header支持多行,缓存已经读取到的数据
	chunked  bool             // 标识是否是chunk
	length   int64            // 数据长度
	streamId uint32           //
}

func (d *decoder) reset() {
	d.state = stateIdent
	d.ident = nil
	d.header = nil
	d.trailer = nil
	d.line = nil
	d.chunked = false
	d.length = 0
	d.streamId = 0
}

func (d *decoder) newEndFrame(ftype netx.FrameType, ident *netx.Identifier, header netx.Header, payload bytex.Buffer) netx.Frame {
	f := netx.NewFrame(ftype, true, d.streamId, ident, header, payload)
	d.reset()
	return f
}

// Decode 不同于官方解析,这里的解析是有状态且非阻塞
// 如果是普通消息,会解析完完整消息后返回Frame,并设置EndFlag为true
// 如果是分块消息,则会分成多帧返回,首帧只包含header,不含payload
func (d *decoder) Decode(buf bytex.Buffer) (netx.Frame, error) {
	for {
		switch d.state {
		case stateIdent:
			if err := d.parseIdent(buf); err != nil {
				return nil, err
			}
		case stateHeader:
			if err := d.parseHeader(buf); err != nil {
				return nil, err
			}
		case stateBody:
			return d.parseBody(buf)
		case stateChunk:
			return d.parseChunk(buf)
		case stateTrailer:
			return d.parseTrailer(buf)
		default:
			panic("invalid state")
		}
	}
}

func (d *decoder) parseIdent(buf bytex.Buffer) error {
	line := buf.ReadLine()
	if line == nil {
		return io.EOF
	}

	tokens := strings.SplitN(line.String(), " ", 3)
	if len(tokens) != 3 {
		return errInvalidHttpIdent
	}

	s1 := tokens[0]
	s2 := tokens[1]
	s3 := tokens[2]

	if !strings.HasPrefix(s1, "HTTP") {
		// parse request identify
		// GET /foo HTTP/1.1
		method := netx.ParseMethod(s1)
		if method == netx.MethodUnknown {
			return errInvalidHttpMethod
		}
		major, minor, ok := http.ParseHTTPVersion(s3)
		if !ok {
			return errInvalidHttpVersion
		}
		ident := &netx.Identifier{
			IsResponse: false,
			IsOneway:   false,
			Version:    toVersion(major, minor),
			URI:        s2,
			Method:     method,
		}
		d.ident = ident
	} else {
		// parse response identify
		// HTTP/1.1 200 OK
		major, minor, ok := http.ParseHTTPVersion(s3)
		if !ok {
			return errInvalidHttpVersion
		}

		code, err := strconv.Atoi(s2)
		if err != nil {
			return fmt.Errorf("parse status code fail, %+v", err)
		}
		ident := &netx.Identifier{
			IsResponse: true,
			Version:    toVersion(major, minor),
			StatusCode: int32(code),
			StatusInfo: s3,
		}
		d.ident = ident
	}

	d.state = stateHeader
	d.header = make(netx.Header, headerSize)
	return nil
}

func (d *decoder) parseHeader(buf bytex.Buffer) error {
	for {
		line, err := d.readContinuedLine(buf)
		if err != nil {
			return err
		}
		if line == nil {
			return io.EOF
		}
		if line.Len() == 0 {
			// read header finish
			d.state = stateBody
			return d.fixHeader()
		}

		kv := line.Bytes()
		if err := setHeader(d.header, kv); err != nil {
			return err
		}
	}
}

func (d *decoder) parseTrailer(buf bytex.Buffer) (netx.Frame, error) {
	if len(d.trailer) == 0 {
		return d.newEndFrame(netx.FrameTypeData, nil, nil, nil), nil
	}

	//
	for {
		line, err := d.readContinuedLine(buf)
		if err != nil {
			return nil, err
		}
		if line == nil {
			return nil, io.EOF
		}
		if line.Len() == 0 {
			return d.newEndFrame(netx.FrameTypeTrailer, nil, d.trailer, nil), nil
		}

		kv := line.Bytes()
		if err := setHeader(d.trailer, kv); err != nil {
			return nil, err
		}
	}
}

func (d *decoder) parseBody(buf bytex.Buffer) (netx.Frame, error) {
	// Prepare body reader. ContentLength < 0 means chunked encoding
	// or close connection when finished, since multipart is not supported yet
	switch {
	case d.chunked:
		d.state = stateChunk
		return netx.NewFrame(netx.FrameTypeHeader, false, d.streamId, d.ident, d.header, nil), nil
	case d.length == 0:
		return d.newEndFrame(netx.FrameTypeHeader, d.ident, d.header, nil), nil
	case d.length > 0:
		payload := buf.ReadN(int(d.length))
		if payload == nil {
			return nil, io.EOF
		}
		return d.newEndFrame(netx.FrameTypeHeader, d.ident, d.header, payload), nil
	default:
		// length < 0, i.e. "Content-Length" not mentioned in header
		return nil, errNoContentLength
	}
}

func (d *decoder) parseChunk(buf bytex.Buffer) (netx.Frame, error) {
	if d.length == -1 {
		line := buf.ReadLine()
		if line == nil {
			return nil, nil
		}

		if line.Len() > maxLineLength {
			return nil, errLineTooLong
		}

		p := line.Bytes()
		p = trimTrailingWhitespace(p)
		p, err := removeChunkExtension(p)
		if err != nil {
			return nil, err
		}
		n, err := parseHexUint(p)
		if err != nil {
			return nil, err
		}
		d.length = int64(n)

		if d.length == 0 {
			d.state = stateTrailer
			return d.parseTrailer(buf)
		}
	}

	chunk := buf.ReadN(int(d.length))
	if chunk != nil {
		return netx.NewFrame(netx.FrameTypeData, false, d.streamId, nil, nil, chunk), nil
	}

	return nil, io.EOF
}

// readContinuedLine 读取多行value
// Historically, HTTP header field values could be extended over
// multiple lines by preceding each extra line with at least one space
// or horizontal tab (obs-fold). This specification deprecates such
// line folding except within the message/http media type
// (Section 8.3.1). A sender MUST NOT generate a message that includes
// line folding (i.e., that has any field-value that contains a match to
// the obs-fold rule) unless the message is intended for packaging
// within the message/http media type.
func (d *decoder) readContinuedLine(buf bytex.Buffer) (bytex.Buffer, error) {
	if d.line == nil {
		// read first line
		line := buf.ReadLine()
		if line == nil || line.Empty() {
			return line, nil
		}

		// Optimistically assume that we have started to buffer the next line
		// and it starts with an ASCII letter (the next header key), or a blank
		// line, so we can avoid copying that buffered data around in memory
		// and skipping over non-existent whitespace.
		ch, err := buf.PeekByte()
		if err == nil && !isContinuedLine(ch) {
			return line, nil
		}

		d.line = line
	}

	// read multi-line
	for {
		ch, err := buf.PeekByte()
		if err != nil {
			return nil, nil
		}
		// end
		if !isContinuedLine(ch) {
			line := d.line
			d.line = nil
			return line, nil
		}

		line := trim(buf.ReadLine().Bytes())
		if len(line) == 0 {
			continue
		}
		_ = d.line.WriteByte(' ')
		_ = d.line.Append(line)
	}
}

func (d *decoder) fixHeader() error {
	if err := d.fixTransferEncoding(); err != nil {
		return err
	}

	if cl, err := d.fixContentLength(); err != nil {
		return err
	} else {
		d.length = cl
	}

	if trailer, err := d.fixTrailer(); err != nil {
		return err
	} else {
		d.trailer = trailer
	}

	// If there is no Content-Length or chunked Transfer-Encoding on a *Response
	// and the status is not 1xx, 204 or 304, then the body is unbounded.
	// See RFC 7230, section 3.3.
	// switch msg.(type) {
	// case *Response:
	// 	if realLength == -1 && !t.Chunked && bodyAllowedForStatus(t.StatusCode) {
	// 		// Unbounded body.
	// 		t.Close = true
	// 	}
	// }

	contentType := d.header.Get(HeaderContentType)
	contentType = removeExtension(contentType)
	d.ident.Codec = uint32(netx.GetCodecType(contentType))

	return nil
}

func (d *decoder) fixTransferEncoding() error {
	raw := d.header.Values(HeaderTransferEncoding)
	if len(raw) == 0 {
		return nil
	}

	d.header.Del(HeaderTransferEncoding)

	// Issue 12785; ignore Transfer-Encoding on HTTP/1.0 requests.
	if !d.protoAtLeast(Version11) {
		return nil
	}

	// Like nginx, we only support a single Transfer-Encoding header field, and
	// only if set to "chunked". This is one of the most security sensitive
	// surfaces in HTTP/1.1 due to the risk of request smuggling, so we keep it
	// strict and simple.
	if len(raw) != 1 {
		return fmt.Errorf("too many transfer encodings: %q", raw)
	}
	if strings.ToLower(textproto.TrimString(raw[0])) != transferEncodingChunked {
		return fmt.Errorf("unsupported transfer encoding: %q", raw[0])
	}

	// RFC 7230 3.3.2 says "A sender MUST NOT send a Content-Length header field
	// in any message that contains a Transfer-Encoding header field."
	//
	// but also: "If a message is received with both a Transfer-Encoding and a
	// Content-Length header field, the Transfer-Encoding overrides the
	// Content-Length. Such a message might indicate an attempt to perform
	// request smuggling (Section 9.5) or response splitting (Section 9.4) and
	// ought to be handled as an error. A sender MUST remove the received
	// Content-Length field prior to forwarding such a message downstream."
	//
	// Reportedly, these appear in the wild.
	d.header.Del(HeaderContentLength)

	d.chunked = true
	return nil
}

func (d *decoder) fixContentLength() (int64, error) {
	isRequest := !d.ident.IsResponse
	contentLens := d.header.Values(HeaderContentLength)
	// Hardening against HTTP request smuggling
	if len(contentLens) > 1 {
		// Per RFC 7230 Section 3.3.2, prevent multiple
		// Content-Length headers if they differ in value.
		// If there are dups of the value, remove the dups.
		// See Issue 16490.
		first := textproto.TrimString(contentLens[0])
		for _, ct := range contentLens[1:] {
			if first != textproto.TrimString(ct) {
				return 0, fmt.Errorf("http: message cannot contain multiple Content-Length headers; got %q", contentLens)
			}
		}

		// deduplicate Content-Length
		d.header.Del(HeaderContentLength)
		d.header.Add(HeaderContentLength, first)

		contentLens = d.header.Values(HeaderContentLength)
	}

	if d.ident.Method == netx.MethodHead {
		// For HTTP requests, as part of hardening against request
		// smuggling (RFC 7230), don't allow a Content-Length header for
		// methods which don't permit bodies. As an exception, allow
		// exactly one Content-Length header if its value is "0".
		if isRequest && len(contentLens) > 0 && !(len(contentLens) == 1 && contentLens[0] == "0") {
			return 0, fmt.Errorf("http: method cannot contain a Content-Length; got %q", contentLens)
		}

		return 0, nil
	}

	if d.ident.IsResponse {
		status := d.ident.StatusCode
		if status/100 == 1 {
			return 0, nil
		}
		switch status {
		case 204, 304:
			return 0, nil
		}
	}

	if d.chunked {
		return -1, nil
	}

	// Logic based on Content-Length
	var cl string
	if len(contentLens) == 1 {
		cl = textproto.TrimString(contentLens[0])
	}
	if cl != "" {
		n, err := parseContentLength(cl)
		if err != nil {
			return -1, err
		}
		return n, nil
	}
	d.header.Del("Content-Length")

	if isRequest {
		// RFC 7230 neither explicitly permits nor forbids an
		// entity-body on a GET request so we permit one if
		// declared, but we default to 0 here (not -1 below)
		// if there's no mention of a body.
		// Likewise, all other request methods are assumed to have
		// no body if neither Transfer-Encoding chunked nor a
		// Content-Length are set.
		return 0, nil
	}

	// Body-EOF logic based on other methods (like closing, or chunked coding)
	return -1, nil
}

// Parse the trailer header
func (d *decoder) fixTrailer() (netx.Header, error) {
	vv := d.header.Values(HeaderTrailer)
	if len(vv) == 0 {
		return nil, nil
	}

	if !d.chunked {
		// Trailer and no chunking:
		// this is an invalid use case for trailer header.
		// Nevertheless, no error will be returned and we
		// let users decide if this is a valid HTTP message.
		// The Trailer header will be kept in Response.Header
		// but not populate Response.Trailer.
		// See issue #27197.
		return nil, nil
	}
	d.header.Del(HeaderTrailer)

	trailer := netx.NewHeader()
	var err error
	for _, v := range vv {
		foreachHeaderElement(v, func(key string) {
			key = http.CanonicalHeaderKey(key)
			switch key {
			case "Transfer-Encoding", "Trailer", "Content-Length":
				if err == nil {
					err = fmt.Errorf("bad trailer key, %+v", key)
					return
				}
			}
			trailer.Set(key, "")
		})
	}
	if err != nil {
		return nil, err
	}
	if len(trailer) == 0 {
		return nil, nil
	}
	return trailer, nil
}

func (d *decoder) protoAtLeast(version uint) bool {
	return d.ident.Version >= version
}

func isContinuedLine(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	n := len(s)
	for n > i && (s[n-1] == ' ' || s[n-1] == '\t') {
		n--
	}
	return s[i:n]
}

// parseContentLength trims whitespace from s and returns -1 if no value
// is set, or the value if it's >= 0.
func parseContentLength(cl string) (int64, error) {
	cl = textproto.TrimString(cl)
	if cl == "" {
		return -1, nil
	}
	n, err := strconv.ParseUint(cl, 10, 63)
	if err != nil {
		return 0, fmt.Errorf("bad Content-Length, %s", cl)
	}
	return int64(n), nil
}

// foreachHeaderElement splits v according to the "#rule" construction
// in RFC 7230 section 7 and calls fn for each non-empty element.
func foreachHeaderElement(v string, fn func(string)) {
	v = textproto.TrimString(v)
	if v == "" {
		return
	}
	if !strings.Contains(v, ",") {
		fn(v)
		return
	}
	for _, f := range strings.Split(v, ",") {
		if f = textproto.TrimString(f); f != "" {
			fn(f)
		}
	}
}

func trimTrailingWhitespace(b []byte) []byte {
	for len(b) > 0 && isASCIISpace(b[len(b)-1]) {
		b = b[:len(b)-1]
	}
	return b
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// removeChunkExtension removes any chunk-extension from p.
// For example,
//     "0" => "0"
//     "0;token" => "0"
//     "0;token=val" => "0"
//     `0;token="quoted string"` => "0"
func removeChunkExtension(p []byte) ([]byte, error) {
	semi := bytes.IndexByte(p, ';')
	if semi == -1 {
		return p, nil
	}
	// TODO: care about exact syntax of chunk extensions? We're
	// ignoring and stripping them anyway. For now just never
	// return an error.
	return p[:semi], nil
}

func removeExtension(s string) string {
	semi := strings.IndexByte(s, ';')
	if semi == -1 {
		return s
	}
	return s[:semi]
}

func parseHexUint(v []byte) (n uint64, err error) {
	for i, b := range v {
		switch {
		case '0' <= b && b <= '9':
			b = b - '0'
		case 'a' <= b && b <= 'f':
			b = b - 'a' + 10
		case 'A' <= b && b <= 'F':
			b = b - 'A' + 10
		default:
			return 0, errors.New("invalid byte in chunk length")
		}
		if i == 16 {
			return 0, errors.New("http chunk length too large")
		}
		n <<= 4
		n |= uint64(b)
	}
	return
}

func setHeader(header netx.Header, kv []byte) error {
	// Key ends at first colon.
	i := bytes.IndexByte(kv, ':')
	if i < 0 {
		return errInvaidHttpHeader
	}
	// key := textproto.CanonicalMIMEHeaderKey(string(kv[:i]))
	key := string(kv[:i])
	// As per RFC 7230 field-name is a token, tokens consist of one or more chars.
	// We could return a ProtocolError here, but better to be liberal in what we
	// accept, so if we get an empty key, skip it.
	if key == "" {
		return nil
	}
	// Skip initial spaces in value.
	i++ // skip colon
	for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
		i++
	}
	value := string(kv[i:])
	// TODO: Optimistically
	// vv := d.header[key]
	// if vv == nil && len(strs) > 0 {
	// 	// More than likely this will be a single-element key.
	// 	// Most headers aren't multi-valued.
	// 	// Set the capacity on strs[0] to 1, so any future append
	// 	// won't extend the slice into the other strings.
	// 	vv, strs = strs[:1:1], strs[1:]
	// 	vv[0] = value
	// 	m[key] = vv
	// } else {
	// 	m[key] = append(vv, value)
	// }
	header.Add(key, value)
	return nil
}
