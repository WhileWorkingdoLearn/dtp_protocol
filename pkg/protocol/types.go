package dtp

import (
	"net"
	"sync"
	"time"
)

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

type Message struct {
	Session    int
	Ip         *net.UDPAddr
	Sender     string
	DataType   string
	DataLength int
	Data       []byte
}

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

type Session struct {
	id              int
	state           State
	expiration      time.Time
	remoteAddr      *net.UDPAddr
	createdAt       time.Time
	lastReceived    time.Time
	lastSend        time.Time
	expiresAt       time.Time
	expectedSeq     uint32
	lastAckedSeq    uint32
	retransmitQueue []struct {
		Sid int
		Msg int
		Pid int
		Bid int
		Lid int
		Pyl []byte
		Rma *net.UDPAddr
	}

	ackTimeout    time.Duration
	authToken     string
	encryptionKey []byte
	customData    map[string]interface{}
}

type SessionHandler struct {
	sessionCache map[int]*Session
	mux          sync.Mutex
}
