package session

import (
	"fmt"
	"sync"
	"time"
)

type SessionHandler struct {
	sessionCache map[int]*Session
	mux          sync.Mutex
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{sessionCache: map[int]*Session{}}
}

const idLength = 4

func (sh *SessionHandler) HasSession(sessionId int) bool {
	_, ok := sh.sessionCache[sessionId]
	return ok
}

func (sh *SessionHandler) GetSession(sessionId int) (*Session, bool) {
	defer sh.mux.Unlock()
	sh.mux.Lock()
	session, ok := sh.sessionCache[sessionId]
	return session, ok
}

func (sh *SessionHandler) NewSession(sessionId int) (*Session, error) {
	defer sh.mux.Unlock()
	sh.mux.Lock()
	_, ok := sh.sessionCache[sessionId]
	if ok {
		return nil, fmt.Errorf("sessionHandler - StartSession, sessionId %v already exists", sessionId)
	}

	newSession := Session{id: sessionId, createdAt: time.Now()}
	sh.sessionCache[sessionId] = &newSession
	return &newSession, nil
}

func (sh *SessionHandler) AddSession(session *Session) error {
	err := session.Validate()
	if err != nil {
		return err
	}

	defer sh.mux.Unlock()
	sh.mux.Lock()

	_, ok := sh.sessionCache[session.id]
	if ok {
		return fmt.Errorf("sessionHandler - StartSession, sessionId %v already exists", &session.id)
	}

	sh.sessionCache[session.id] = session

	return nil
}

func (sh *SessionHandler) RemoveSession(sessionId int) error {
	defer sh.mux.Unlock()
	sh.mux.Lock()
	_, ok := sh.sessionCache[sessionId]
	if ok {
		delete(sh.sessionCache, sessionId)
		return nil
	}
	return fmt.Errorf("Session not found %v", sessionId)
}
