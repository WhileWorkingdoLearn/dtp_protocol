package session

import (
	"net"
	"time"
)

type State int

const (
	Open State = iota
	Closed
)

type SessionHandler struct {
	id           string
	state        State
	expiration   time.Time
	remoteAddr   *net.UDPAddr
	createdAt    time.Time
	lastReceived time.Time
	lastSend     time.Time
	expiresAt    time.Time
	expectedSeq  uint32
	lastAckedSeq uint32
	//retransmitQueue []Packet
	ackTimeout time.Duration

	authToken     string
	encryptionKey []byte

	customData map[string]interface{}
}

func (s SessionHandler) ChangeState(newState State) {
	s.state = newState
}
