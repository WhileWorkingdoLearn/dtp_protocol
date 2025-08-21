package transport

import (
	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
)

// UDP transport stubs. Helpful if your protocol is message-oriented.
type udpTransport struct{}

func NewUDP() Transport { return &udpTransport{} }

func (t *udpTransport) Dial(addr string, opts dtp.Options) (dtp.Conn, error) {
	// TODO: wrap *net.UDPConn into a packet-oriented Conn
	_ = addr
	_ = opts
	return nil, nil
}

func (t *udpTransport) Listen(addr string, opts dtp.Options) (dtp.Listener, error) {
	// TODO: wrap *net.UDPConn in a Listener-like acceptor (pseudo-Conn per remote)
	_ = addr
	_ = opts
	return nil, nil
}
