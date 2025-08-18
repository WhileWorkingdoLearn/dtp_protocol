package session

import (
	"net"
	"time"

	"github.com/WhilecodingDoLearn/dtp/protocol"
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

func (s *Session) ChangeState(newState int) {
	s.state = newState

}

func (s *Session) IsAlive() bool {
	return s.state != protocol.ALI
}

func (s *Session) IsOpen() bool {
	return s.state != protocol.OPN
}

func (s *Session) IsPending() bool {
	return s.state != protocol.RTY
}
