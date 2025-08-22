package main

import (
	"fmt"
	"testing"

	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	tests := []struct {
		name            string
		p               codec.Package
		expCode         codec.State
		expSend         bool
		expHandlerState codec.State
		expDatasize     int
	}{
		{name: "REQ To large payload Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.REQ", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ, PayloadLength: 2028}, expCode: codec.ERR, expHandlerState: codec.ERR, expSend: true},
		{name: "REQ Connection Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.REQ", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ}, expCode: codec.OPN, expHandlerState: codec.OPN, expSend: true},
		{name: "REQ Another Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.OPEN", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ}, expCode: codec.REQ, expHandlerState: codec.OPN, expSend: false},
		{name: "ACK Request -> p.MsgCode == codec.ACK | connectionHandler.state == codec.OPEN", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.ACK}, expCode: codec.ALI, expHandlerState: codec.ALI, expSend: true},
	}

	connHandler := ConnectionHandler{state: codec.REQ}

	for _, subTest := range tests {
		res, send := handle(subTest.p, &connHandler)
		assert.Equal(t, subTest.expSend, send, fmt.Sprintf("name:%v - %v \n", subTest.name, "test if message should be send"))
		assert.Equal(t, subTest.expCode, res.MSgCode, fmt.Sprintf("name:%v - %v \n", subTest.name, "test response code of message"))
		assert.Equal(t, subTest.expHandlerState, connHandler.state, fmt.Sprintf("name:%v - %v \n", subTest.name, "test state of connectionHandler"))
	}

}

func TestBuffer(t *testing.T) {
	b := Buffer{frames: [1024]byte{}}
	payload := []byte("ABCD")
	p := codec.Package{SessionID: 122, PackedID: 0, FrameBegin: 2, FrameEnd: 1 + len(payload), PayloadLength: len(payload), Payload: payload}
	err := b.Read(p)
	assert.Nil(t, err, "test for error")
	assert.Equal(t, []byte("ABCD"), b.frames[2:5])
}
