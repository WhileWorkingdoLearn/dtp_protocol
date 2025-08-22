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
		{name: "To large payload Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.REQ", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ, PayloadLength: 2028}, expCode: codec.ERR, expHandlerState: codec.ERR, expSend: true},
		{name: "Connection Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.REQ", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ}, expCode: codec.OPN, expHandlerState: codec.OPN, expSend: true},
		{name: "Another Request -> p.MsgCode == codec.REQ | connectionHandler.state == codec.OPEN", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.REQ}, expCode: codec.REQ, expHandlerState: codec.OPN, expSend: false},
		{name: "Ack Request -> p.MsgCode == codec.ACK | connectionHandler.state == codec.OPEN", p: codec.Package{SessionID: 1234, UserID: 111, MSgCode: codec.ACK}, expCode: codec.ALI, expHandlerState: codec.ALI, expSend: true},
	}

	connHandler := ConnectionHandler{state: codec.REQ}

	for _, subTest := range tests {
		res, send := handle(subTest.p, &connHandler)
		assert.Equal(t, subTest.expSend, send, fmt.Sprintf("name:%v - %v \n", subTest.name, "test if message should be send"))
		assert.Equal(t, subTest.expCode, res.MSgCode, fmt.Sprintf("name:%v - %v \n", subTest.name, "test response code of message"))
		assert.Equal(t, subTest.expHandlerState, connHandler.state, fmt.Sprintf("name:%v - %v \n", subTest.name, "test state of connectionHandler"))
	}

}
