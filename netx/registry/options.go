package registry

import (
	"crypto/tls"
	"time"
)

type Options struct {
	Addrs   []string
	Timeout time.Duration
	Secure  bool
	TLS     *tls.Config
}

type Option func(o *Options)
