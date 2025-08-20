package dtp

import "net"

type State int

const (
	REQ State = iota
	OPN
	ALI
	CLD
	ACK
	RTY
	ERR
)

type Package struct {
	Sid int
	Msg State
	Pid int
	Bid int
	Lid int
	Pyl []byte
	Rma *net.UDPAddr
}

type PackageReader interface {
	Read() Package
}
