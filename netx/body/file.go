package body

import (
	"fmt"
	"io"
	"os"

	"github.com/foredata/nova/pkg/bytex"
)

// NewFileBody 用于传输文件,可以指定文件偏移offset和每次传输chunck大小
func NewFileBody(filename string, offset int64, maxChunk int) (Body, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("invalid file body")
	}

	size := fi.Size()
	if size <= offset {
		return NewNoBody(), nil
	}

	b := &fileBody{file: file, leftSize: size - offset, maxChunk: int64(maxChunk)}
	return b, nil
}

// fileBody 文件类型
type fileBody struct {
	noWriter
	file     *os.File
	leftSize int64
	maxChunk int64
}

func (b *fileBody) End() bool {
	return b.leftSize == 0
}

func (b *fileBody) Close() error {
	if b.file != nil {
		err := b.file.Close()
		b.file = nil
		return err
	}

	return nil
}

func (b *fileBody) Read(p []byte) (int, error) {
	if b.file == nil {
		return 0, io.EOF
	}

	n, err := b.file.Read(p)
	if err == io.EOF {
		_ = b.file.Close()
		b.file = nil
	}
	return n, err
}

func (b *fileBody) ReadFast(blocking bool) (bytex.Buffer, error) {
	if b.file == nil {
		return nil, io.EOF
	}

	var p []byte
	if b.leftSize < int64(b.maxChunk) {
		p = make([]byte, int(b.leftSize))
	} else {
		p = make([]byte, int(b.maxChunk))
	}
	n, err := b.file.Read(p)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		_ = b.file.Close()
		b.file = nil
		p = p[:n]
	}
	buff := bytex.NewBuffer()
	_ = buff.Append(p)
	return buff, err
}

func (b *fileBody) Buffer() (bytex.Buffer, error) {
	return nil, ErrNotSupport
}
