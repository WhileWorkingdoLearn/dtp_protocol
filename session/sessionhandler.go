package session

import (
	"fmt"
	"sync"

	"github.com/WhilecodingDpLearn/dtp/protocol"
)

type SessionHandler struct {
	sessionCache map[string]Session
	mux          sync.Mutex
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{sessionCache: map[string]Session{}}
}

const idLength = 4

func (sh *SessionHandler) validatePackage(p protocol.Package) error {
	if len(p.SessionId) != idLength || len(p.Id) != idLength {
		return fmt.Errorf("invalid package: missing session id or packageid")
	}
	return nil
}

func (sh *SessionHandler) Handle(p protocol.Package) error {
	defer sh.mux.Unlock()

	sh.mux.Lock()

	err := sh.validatePackage(p)
	if err != nil {
		return err
	}

	session, ok := sh.sessionCache[p.SessionId]
	if ok {
		session.Receive(p)
		return nil
	}

	session = NewSession()
	session.Receive(p)

	sh.sessionCache[p.SessionId] = session
	return nil
}

func (sh *SessionHandler) HasSession(sessionId string) bool {
	_, ok := sh.sessionCache[sessionId]
	return ok
}
