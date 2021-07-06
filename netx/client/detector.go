package client

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

type detector struct {
	protocol netx.Protocol
}

func (d *detector) Detect(p bytex.Peeker) netx.Protocol {
	return nil
}

func (d *detector) Default() netx.Protocol {
	return d.protocol
}
