package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionHandler(t *testing.T) {
	sessionHandler := NewSessionHandler()
	session, err := sessionHandler.NewSession(123456)
	assert.NotNil(t, session)
	assert.Nil(t, err)

	errAdd := sessionHandler.AddSession(session)
	assert.NotNil(t, errAdd)

	has := sessionHandler.HasSession(123456)
	assert.Equal(t, true, has)

	sessionInStore := sessionHandler.GetSession(123456)
	assert.NotNil(t, sessionInStore)
	assert.Equal(t, 123456, sessionInStore.id)

	errRem := sessionHandler.RemoveSession(123456)
	assert.Nil(t, errRem)

	hasNot := sessionHandler.HasSession(123456)
	assert.Equal(t, false, hasNot)

}
