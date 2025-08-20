package dtp

import (
	"net"
	"time"
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

type Session struct {
	id              int
	state           int
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

func NewSession() Session {
	return Session{}
}

func (sh *Session) Validate() error {
	return nil
}

func (s *Session) ChangeState(newState int) {
	s.state = newState

}

func (s *Session) IsAlive() bool {
	return s.state != ALI
}

func (s *Session) IsOpen() bool {
	return s.state != OPN
}

func (s *Session) IsPending() bool {
	return s.state != RTY
}
