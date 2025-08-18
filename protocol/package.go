package protocol

import (
	"net"
)

const (
	REQ int = iota
	OPN
	ALI
	CLD
	ACK
	RTY
	ERR
)

type Package struct {
	Sid int
	Msg int
	Pid int
	Bid int
	Lid int
	Pyl []byte
	Rma *net.UDPAddr
}

type PackageReader interface {
	Read() Package
}
