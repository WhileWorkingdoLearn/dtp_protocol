package dtp

import (
	"net"
	"sync"
)

type Frame struct {
	start int
	end   int
}

type PacketCache struct {
	cache []struct {
		Sid int
		Msg int
		Pid int
		Bid int
		Lid int
		Pyl []byte
		Rma *net.UDPAddr
	}
	received int
}

type Buffer struct {
	frames     map[Frame]*PacketCache
	buffer     []string
	bufferSize int
	mux        sync.Mutex
}
