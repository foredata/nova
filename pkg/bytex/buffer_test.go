package bytex

import (
	"io"
	"testing"
)

func TestBuffer(t *testing.T) {
	SetChunkSize(4)
	b := newBuffer()
	_, _ = b.Write([]byte("aaaa"))
	_, _ = b.Write([]byte("bbbb"))
	// assert.Equal(b.Len(), 8)
	_, _ = b.Seek(0, io.SeekStart)

	p1 := make([]byte, 6)
	_, _ = b.Peek(p1)
	t.Log(p1)

	r1 := make([]byte, 3)
	_, _ = b.Read(r1)
	t.Log(string(r1))
	r2 := make([]byte, 3)
	_, _ = b.Read(r2)
	t.Log(string(r2))
	r3 := make([]byte, 3)
	_, err := b.Read(r3)
	t.Log(err, "\n")
}

func TestReadLine(t *testing.T) {
	data := `
	Connection: keep-alive
	Pragma: no-cache
	Cache-Control: no-cache
	sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"
	User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36
	sec-ch-ua-mobile: ?0
	Accept: */*
	Sec-Fetch-Site: same-origin
	Sec-Fetch-Mode: cors
	Sec-Fetch-Dest: empty
	Accept-Encoding: gzip, deflate, br
	Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
	`
	b := newBuffer()
	_ = b.Append(data)
	for {
		line := b.ReadLine()
		if line == nil {
			break
		}
		t.Log(line.String())
	}
}
