package session

import (
	"testing"

	"github.com/WhilecodingDpLearn/dtp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestSessionHandler(t *testing.T) {
	sh := NewSessionHandler()
	err := sh.Handle(protocol.Package{SessionId: "Hell", Id: "1234", Data: "Data"})
	assert.Nil(t, err)
	has := sh.HasSession("Hell")
	assert.Equal(t, true, has)
}
