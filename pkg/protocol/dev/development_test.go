package main

import (
	"fmt"
	"testing"

	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name   string
		pkg    codec.Package
		expErr *dtp.PacketError
	}{
		{
			name: "wrong sessions id",
			pkg:  codec.Package{SessionID: -1, PackedID: 123},
			expErr: &dtp.PacketError{
				Text:     "wrong sessions id",
				Want:     0,
				Has:      -1,
				PacketID: 123,
			},
		},
		{
			name: "illigal packet state",
			pkg:  codec.Package{SessionID: 1, PackedID: 123, MSgCode: codec.ALI},
			expErr: &dtp.PacketError{
				Text:     "illigal packet state",
				Want:     int(codec.REQ),
				Has:      int(codec.ALI),
				PacketID: 123,
			},
		},
		{
			name:   "packet state RTY",
			pkg:    codec.Package{SessionID: 1, PackedID: 123, MSgCode: codec.RTY},
			expErr: nil,
		},
		{
			name:   "packet state ERR",
			pkg:    codec.Package{SessionID: 1, PackedID: 123, MSgCode: codec.ERR},
			expErr: nil,
		},
		{
			name: "illigal packet id",
			pkg:  codec.Package{SessionID: 1, PackedID: -1, MSgCode: codec.REQ},
			expErr: &dtp.PacketError{
				Text:     "illigal packet id",
				Want:     0,
				Has:      -1,
				PacketID: -1,
			},
		},
		{
			name: "frame begin negative",
			pkg: codec.Package{
				SessionID:  1,
				PackedID:   123,
				MSgCode:    codec.REQ,
				FrameBegin: -1,
			},
			expErr: &dtp.PacketError{
				Text:     "frame begin out of range",
				Want:     0,
				Has:      -1,
				PacketID: 123,
			},
		},
		{
			name: "frame begin too large",
			pkg: codec.Package{
				SessionID:  1,
				PackedID:   123,
				MSgCode:    codec.REQ,
				FrameBegin: BufferSize + 1,
			},
			expErr: &dtp.PacketError{
				Text:     "frame begin out of range",
				Want:     0,
				Has:      BufferSize + 1,
				PacketID: 123,
			},
		},
		{
			name: "frame end negative",
			pkg: codec.Package{
				SessionID:  1,
				PackedID:   123,
				MSgCode:    codec.REQ,
				FrameBegin: 0,
				FrameEnd:   -1,
			},
			expErr: &dtp.PacketError{
				Text:     "frame end out of range",
				Want:     BufferSize,
				Has:      -1,
				PacketID: 123,
			},
		},
		{
			name: "frame end too large",
			pkg: codec.Package{
				SessionID:  1,
				PackedID:   123,
				MSgCode:    codec.REQ,
				FrameBegin: 0,
				FrameEnd:   BufferSize + 1,
			},
			expErr: &dtp.PacketError{
				Text:     "frame end out of range",
				Want:     BufferSize,
				Has:      BufferSize + 1,
				PacketID: 123,
			},
		},
		{
			name: "payload length negative",
			pkg: codec.Package{
				SessionID:     1,
				PackedID:      123,
				MSgCode:       codec.REQ,
				FrameBegin:    0,
				FrameEnd:      0,
				PayloadLength: -1,
			},
			expErr: &dtp.PacketError{
				Text:     "invalid payload size",
				Want:     BufferSize,
				Has:      -1,
				PacketID: 123,
			},
		},
		{
			name: "payload length too large",
			pkg: codec.Package{
				SessionID:     1,
				PackedID:      123,
				MSgCode:       codec.REQ,
				FrameBegin:    0,
				FrameEnd:      0,
				PayloadLength: BufferSize + 1,
			},
			expErr: &dtp.PacketError{
				Text:     "invalid payload size",
				Want:     BufferSize,
				Has:      BufferSize + 1,
				PacketID: 123,
			},
		},
		{
			name: "corrupt payload size",
			pkg: codec.Package{
				SessionID:     1,
				PackedID:      123,
				MSgCode:       codec.REQ,
				FrameBegin:    0,
				FrameEnd:      0,
				PayloadLength: 3,
				Payload:       []byte{0x01, 0x02},
			},
			expErr: &dtp.PacketError{
				Text:     "corrupt payload size",
				Want:     3,
				Has:      2,
				PacketID: 123,
			},
		},
		{
			name: "valid packet",
			pkg: codec.Package{
				SessionID:     1,
				PackedID:      42,
				MSgCode:       codec.REQ,
				FrameBegin:    0,
				FrameEnd:      BufferSize,
				PayloadLength: 3,
				Payload:       []byte{0xA, 0xB, 0xC},
			},
			expErr: nil,
		},
	}

	connHandler := DTPConnection{state: codec.REQ, sessionId: 1}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := connHandler.validate(tc.pkg)

			if tc.expErr != nil {
				assert.Error(t, err)
				pe, ok := err.(*dtp.PacketError)
				assert.True(t, ok, "expected a *PacketError")

				assert.Equal(t, tc.expErr.Text, pe.Text, "Text")
				assert.Equal(t, tc.expErr.Want, pe.Want, "Want")
				assert.Equal(t, tc.expErr.Has, pe.Has, "Has")
				assert.Equal(t, tc.expErr.PacketID, pe.PacketID, "PacketID")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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

	connHandler := DTPConnection{state: codec.REQ}

	for _, subTest := range tests {
		t.Run(subTest.name, func(t *testing.T) {
			res, send := connHandler.handle(subTest.p)
			assert.Equal(t, subTest.expSend, send, fmt.Sprintf("name:%v - %v \n", subTest.name, "test if message should be send"))
			assert.Equal(t, subTest.expCode, res.MSgCode, fmt.Sprintf("name:%v - %v \n", subTest.name, "test response code of message"))
			assert.Equal(t, subTest.expHandlerState, connHandler.state, fmt.Sprintf("name:%v - %v \n", subTest.name, "test state of connectionHandler"))
		})
	}

}

func TestBuffer_Read(t *testing.T) {
	buf := &Buffer{}

	tests := []struct {
		name   string
		pkg    codec.Package
		expErr string
	}{
		{
			name:   "frame begin negative",
			pkg:    codec.Package{FrameBegin: -1, FrameEnd: 0, Payload: []byte{0}},
			expErr: "frame begin index -1 out of range [0:1023]",
		},
		{
			name:   "frame begin too large",
			pkg:    codec.Package{FrameBegin: BufferSize, FrameEnd: BufferSize, Payload: []byte{0}},
			expErr: fmt.Sprintf("frame begin index %d out of range [0:%d]", BufferSize, BufferSize-1),
		},
		{
			name:   "frame end negative",
			pkg:    codec.Package{FrameBegin: 0, FrameEnd: -1, Payload: []byte{0}},
			expErr: "frame end index -1 out of range [0:1023]",
		},
		{
			name:   "frame end too large",
			pkg:    codec.Package{FrameBegin: 0, FrameEnd: BufferSize, Payload: []byte{0}},
			expErr: fmt.Sprintf("frame end index %d out of range [0:%d]", BufferSize, BufferSize-1),
		},
		{
			name:   "begin greater than end",
			pkg:    codec.Package{FrameBegin: 10, FrameEnd: 5, Payload: []byte{}},
			expErr: "invalid frame range: begin 10 > end 5",
		},
		{
			name:   "payload length too short",
			pkg:    codec.Package{FrameBegin: 2, FrameEnd: 4, Payload: []byte{1, 2}},
			expErr: "payload length mismatch: got 2 bytes, expected 3",
		},
		{
			name:   "payload length too long",
			pkg:    codec.Package{FrameBegin: 2, FrameEnd: 4, Payload: []byte{1, 2, 3, 4}},
			expErr: "payload length mismatch: got 4 bytes, expected 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := buf.Read(tc.pkg)
			assert.EqualError(t, err, tc.expErr)
		})
	}

	t.Run("valid read", func(t *testing.T) {
		// reset buffer
		buf = &Buffer{}
		payload := []byte{0xA, 0xB, 0xC}
		pkg := codec.Package{
			FrameBegin:    5,
			FrameEnd:      7,
			PayloadLength: len(payload),
			Payload:       payload,
		}

		err := buf.Read(pkg)
		assert.NoError(t, err)

		// verify bytes copied into internal frames slice
		for i, b := range payload {
			assert.Equal(t, b, buf.frames[5+i])
		}
	})
}
