package netx

import (
	"fmt"
	"io"
	"net/http"

	"github.com/foredata/nova/pkg/runtimes"
	"github.com/foredata/nova/pkg/strx"
)

// Error 网络错误,额外提供code和status
type Error interface {
	error
	Code() int
	Status() string
}

// NewError 通过错误码,status创建error
func NewError(code int, status string, format string, args ...interface{}) Error {
	if status == "" {
		status = http.StatusText(code)
	}
	return &netError{
		code:   code,
		status: status,
		detail: strx.Sprintf(format, args...),
		// stack:  callers(),
	}
}

// WrapError .
func WrapError(err error, code int, status string, format string, args ...interface{}) Error {
	return &netError{
		cause:  err,
		code:   code,
		status: status,
		detail: strx.Sprintf(format, args...),
	}
}

type netError struct {
	cause  error
	code   int
	status string
	detail string
	stack  *runtimes.Stack
}

func (e *netError) Code() int {
	return e.code
}

func (e *netError) Status() string {
	return e.status
}

func (e *netError) Unwrap() error {
	return e.cause
}

func (e *netError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d:%s] %s: %s", e.code, e.status, e.detail, e.cause.Error())
	} else {
		return fmt.Sprintf("[%d:%s] %s", e.code, e.status, e.detail)
	}
}

func (e *netError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.cause != nil {
				fmt.Fprintf(s, "%+v\n", e.cause)
			}
			// e.stack.Format(s, verb)
			io.WriteString(s, e.detail)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

// BadRequest generates a 400 error.
func BadRequest(format string, args ...interface{}) error {
	return NewError(http.StatusBadRequest, "", format, args...)
}

// Unauthorized generates a 401 error.
func Unauthorized(format string, args ...interface{}) error {
	return NewError(http.StatusUnauthorized, "", format, args...)
}

// Forbidden generates a 403 error.
func Forbidden(format string, args ...interface{}) error {
	return NewError(http.StatusForbidden, "", format, args...)
}

// NotFound generates a 404 error.
func NotFound(format string, args ...interface{}) error {
	return NewError(http.StatusNotFound, "", format, args...)
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(format string, args ...interface{}) error {
	return NewError(http.StatusMethodNotAllowed, "", format, args...)
}

// RequestTimeout generates a 408 error.
func RequestTimeout(format string, args ...interface{}) error {
	return NewError(http.StatusRequestTimeout, "", format, args...)
}

// Conflict generates a 409 error.
func Conflict(format string, args ...interface{}) error {
	return NewError(http.StatusConflict, "", format, args...)
}

// InternalServerError generates a 500 error.
func InternalServerError(format string, args ...interface{}) error {
	return NewError(http.StatusInternalServerError, "", format, args...)
}

// NotImplemented generates a 501 error.
func NotImplemented(format string, args ...interface{}) error {
	return NewError(http.StatusNotImplemented, "", format, args...)
}

// BadGateway generates a 502 error.
func BadGateway(format string, args ...interface{}) error {
	return NewError(http.StatusBadGateway, "", format, args...)
}

// ServiceUnavailable generates a 503 error.
func ServiceUnavailable(format string, args ...interface{}) error {
	return NewError(http.StatusServiceUnavailable, "", format, args...)
}

// GatewayTimeout generates a 504 error.
func GatewayTimeout(format string, args ...interface{}) error {
	return NewError(http.StatusGatewayTimeout, "", format, args...)
}
