package theader

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

type encoder struct {
}

func (encoder) Encode(conn netx.Conn, frame netx.Frame) (bytex.Buffer, error) {

	return nil, nil
}
