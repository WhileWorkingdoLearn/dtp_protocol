package session

import (
	"net"
	"time"

	"github.com/WhilecodingDpLearn/dtp/protocol"
)

type State int

const (
	Open State = iota
	Pending
	Closed
)

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
	retransmitQueue []protocol.Package
	ackTimeout      time.Duration
	authToken       string
	encryptionKey   []byte
	customData      map[string]interface{}
}

func NewSession() Session {
	return Session{}
}

func (sh *Session) Validate() error {
	return nil
}

func (s *Session) ChangeState(newState State) {
	s.state = newState
}

func (s *Session) IsAlive() bool {
	return s.state != Closed
}

func (s *Session) IsOpen() bool {
	return s.state != Open
}

func (s *Session) IsPending() bool {
	return s.state != Pending
}

func (s *Session) Receive(p protocol.Package) {
	if s.id == 0 {
		s.id = p.Sid
		now := time.Now()
		s.createdAt = now
		s.lastReceived = now
		s.expiresAt = now.Add(60 * time.Second)
	}

}

func (s *Session) Send(p protocol.Package) {

}
