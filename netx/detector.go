package netx

import "github.com/foredata/nova/pkg/bytex"

// NewDetector 使用指定协议
func NewDetector(p Protocol) Detector {
	d := &detector{proto: p}
	return d
}

type detector struct {
	proto Protocol
}

func (d *detector) Detect(p bytex.Peeker) Protocol {
	return d.proto
}

func (d *detector) Default() Protocol {
	return d.proto
}
