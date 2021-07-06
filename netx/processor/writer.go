package processor

import (
	"io"

	"github.com/foredata/nova/netx"
)

type packetWriter struct {
	proto netx.Protocol
}

func (pw *packetWriter) Close() error {
	return nil
}

func (pw *packetWriter) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}
