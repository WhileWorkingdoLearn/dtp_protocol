package dtp

import (
	"net"
	"time"

	dtp "github.com/WhilecodingDoLearn/dtp/protocol/types"
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
	state           dtp.State
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

func (sh *Session) Validate() error {
	return nil
}

func (sh *Session) State() dtp.State {
	return sh.state
}

// Creates a new session
func NewSession(sessionId int) *Session {
	newSession := Session{id: sessionId, createdAt: time.Now(), state: dtp.REQ}
	return &newSession
}
