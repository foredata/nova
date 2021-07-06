package server

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/protocol"
	"github.com/foredata/nova/pkg/bytex"
)

var gDetector = &detector{}

func newDetector() netx.Detector {
	return gDetector
}

// detector 默认自动探测协议
type detector struct {
}

func (d *detector) Detect(p bytex.Peeker) netx.Protocol {
	return protocol.Detect(p)
}

func (d *detector) Default() netx.Protocol {
	return nil
}
