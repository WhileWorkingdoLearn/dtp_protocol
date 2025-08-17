package protocol

import (
	"net"
)

const (
	REQ int = iota
	OPN
	ACK
	RTY
	ERR
)

type Package struct {
	Sid int
	Msg int
	PId int
	Bid int
	Lid int
	Pyl []byte
	Rma *net.UDPAddr
}

type PackageReader interface {
	Read() Package
}
