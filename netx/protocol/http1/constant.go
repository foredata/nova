package http1

import (
	"net/http"

	"github.com/foredata/nova/netx"
)

// Headers
const (
	HeaderContentType      = "Content-Type"
	HeaderContentLength    = "Content-Length"
	HeaderTransferEncoding = "Transfer-Encoding"
	HeaderTrailer          = "Trailer"
)

const (
	Version09 = 9  // HTTP0.9
	Version10 = 10 // HTTP1.0
	Version11 = 11 // HTTP1.1
)

const (
	transferEncodingChunked = "chunked"
)

var (
	kCRLF = []byte("\r\n") // http 分隔符
)

func toHttpVersion(version uint) string {
	switch version {
	case 9:
		return "HTTP/0.9"
	case 10:
		return "HTTP/1.0"
	default:
		return "HTTP/1.1"
	}
}

func toVersion(major, minor int) uint {
	return uint(major*10 + minor)
}

func toHttpStatus(ident *netx.Identifier) (int, string) {
	code := int(ident.StatusCode)
	info := ident.StatusInfo
	if code == 0 {
		code = http.StatusOK
	}
	if info == "" {
		info = http.StatusText(code)
	}

	return code, info
}
