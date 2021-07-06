package errorx

import (
	"errors"
	"fmt"
	"io"

	"github.com/foredata/nova/pkg/runtimes"
	"github.com/foredata/nova/pkg/strx"
)

// some common errors
var (
	ErrNotSupport    = errors.New("not support")
	ErrNotFound      = errors.New("not found")
	ErrInvalidConfig = errors.New("invalid config")
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return &wrapErr{
		msg:   message,
		stack: runtimes.Callers(),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return &wrapErr{
		msg:   strx.Sprintf(format, args...),
		stack: runtimes.Callers(),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) error {
	if err == nil {
		return New(message)
	}

	return &wrapErr{
		cause: err,
		msg:   message,
		stack: runtimes.Callers(),
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return Errorf(format, args...)
	}

	return &wrapErr{
		cause: err,
		msg:   strx.Sprintf(format, args...),
		stack: runtimes.Callers(),
	}
}

type wrapErr struct {
	cause error
	msg   string
	stack *runtimes.Stack
}

func (w *wrapErr) Error() string {
	if w.cause != nil {
		return w.msg + ": " + w.cause.Error()
	} else {
		return w.msg
	}
}

func (w *wrapErr) Unwrap() error {
	return w.cause
}

func (w *wrapErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if w.cause != nil {
				fmt.Fprintf(s, "%+v\n", w.cause)
			}
			w.stack.Format(s, verb)
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
