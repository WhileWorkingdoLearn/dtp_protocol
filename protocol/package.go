package protocol

import "net"

type Package struct {
	SessionId  string
	Id         string
	Data       string
	RemoteAddr *net.UDPAddr
}
