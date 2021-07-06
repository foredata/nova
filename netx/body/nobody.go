package body

import (
	"io"

	"github.com/foredata/nova/pkg/bytex"
)

func NewNoBody() Body {
	return gNoBody
}

var gNoBody = &noBody{}

type noWriter struct {
}

func (d *noWriter) Write(bytex.Buffer) error {
	return ErrNotSupport
}

func (d *noWriter) Flush() {
}

type noBody struct {
	noWriter
}

func (d *noBody) Close() error {
	return nil
}

func (d *noBody) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (d *noBody) ReadFast(blocking bool) (bytex.Buffer, error) {
	return nil, io.EOF
}

func (d *noBody) Buffer() (bytex.Buffer, error) {
	return nil, nil
}

func (d *noBody) End() bool {
	return true
}
