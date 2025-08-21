package transport

import (
	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
)

// Transport abstracts dial/listen for different underlying protocols (TCP/UDP/QUIC).
type Transport interface {
	Dial(addr string, opts dtp.Options) (dtp.Conn, error)
	Listen(addr string, opts dtp.Options) (dtp.Listener, error)
}
