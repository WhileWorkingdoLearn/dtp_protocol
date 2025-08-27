package main

import (
	"testing"
	"time"

	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewPkg(msg codec.State, sID int, pID int) codec.Package {
	return codec.Package{
		MSgCode:       msg,
		SessionID:     sID,
		PackedID:      pID,
		FrameBegin:    0,
		FrameEnd:      0,
		PayloadLength: 0,
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name   string
		pkg    codec.Package
		expErr *dtp.PacketError
	}{
		{
			name: "wrong sessions id",
			pkg:  NewPkg(codec.REQ, -1, 123),
			expErr: &dtp.PacketError{
				Text:     "wrong sessions id",
				Want:     0,
				Has:      -1,
				PacketID: 123,
			},
		},
		{
			name: "illigal packet id",
			pkg:  NewPkg(codec.REQ, 1, -1),
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

// Helpers für Tests

func newConnWithState(s codec.State) *DTPConnection {
	return &DTPConnection{
		prevState:    s,
		state:        s,
		lastReceived: time.Now(),
	}
}

func makePkg(msg codec.State) codec.Package {
	return codec.Package{
		MSgCode:       msg,
		SessionID:     42,
		PackedID:      7,
		FrameBegin:    0,
		FrameEnd:      0,
		PayloadLength: 0,
	}
}

type transitionCase struct {
	name         string
	startState   codec.State
	inMsg        codec.State
	expNextState codec.State
	expOutMsg    *codec.State // nil => nicht prüfen (z. B. bei keepalive/silent)
	expSend      bool
}

// Einzelne Transitionen als Tabelle abprüfen.
func TestFSM_Transitions_Table(t *testing.T) {
	cases := []transitionCase{
		// Happy Path
		{name: "REQ + REQ -> OPN", startState: codec.REQ, inMsg: codec.REQ, expNextState: codec.OPN, expOutMsg: ptr(codec.OPN), expSend: true},
		{name: "OPN + OPN -> ACK", startState: codec.OPN, inMsg: codec.OPN, expNextState: codec.ACK, expOutMsg: ptr(codec.ACK), expSend: true},
		{name: "ACK + ALI -> ALI", startState: codec.ACK, inMsg: codec.ALI, expNextState: codec.ALI, expOutMsg: ptr(codec.ALI), expSend: true},

		// ALI: keepalive (silent)
		{name: "ALI + ALI (keepalive) -> ALI silent", startState: codec.ALI, inMsg: codec.ALI, expNextState: codec.ALI, expOutMsg: nil, expSend: false},

		// ALI: Abzweigungen
		{name: "ALI + RTY -> RTY", startState: codec.ALI, inMsg: codec.RTY, expNextState: codec.RTY, expOutMsg: ptr(codec.RTY), expSend: true},
		{name: "ALI + FIN -> FIN", startState: codec.ALI, inMsg: codec.FIN, expNextState: codec.FIN, expOutMsg: ptr(codec.FIN), expSend: true},
		{name: "ALI + CLD -> CLD", startState: codec.ALI, inMsg: codec.CLD, expNextState: codec.CLD, expOutMsg: ptr(codec.CLD), expSend: true},

		// ALI: unerwartete Message -> ERR
		{name: "ALI + OPN (unexpected) -> ERR", startState: codec.ALI, inMsg: codec.OPN, expNextState: codec.ERR, expOutMsg: ptr(codec.ERR), expSend: true},

		// RTY
		{name: "RTY + ACK (retryOK) -> ALI", startState: codec.RTY, inMsg: codec.ACK, expNextState: codec.ALI, expOutMsg: ptr(codec.ALI), expSend: true},
		{name: "RTY + OPN (fail) -> ERR", startState: codec.RTY, inMsg: codec.OPN, expNextState: codec.ERR, expOutMsg: ptr(codec.ERR), expSend: true},

		// FIN
		{name: "FIN + OPN (reopen) -> OPN", startState: codec.FIN, inMsg: codec.OPN, expNextState: codec.OPN, expOutMsg: ptr(codec.OPN), expSend: true},
		{name: "FIN + FIN (close) -> CLD", startState: codec.FIN, inMsg: codec.FIN, expNextState: codec.CLD, expOutMsg: ptr(codec.CLD), expSend: true},

		// ERR
		{name: "ERR + OPN (recovery) -> OPN", startState: codec.ERR, inMsg: codec.OPN, expNextState: codec.OPN, expOutMsg: ptr(codec.OPN), expSend: true},
		{name: "ERR + FIN (stay err) -> ERR", startState: codec.ERR, inMsg: codec.FIN, expNextState: codec.ERR, expOutMsg: ptr(codec.ERR), expSend: true},

		// CLD terminal
		{name: "CLD + ALI (terminal, silent)", startState: codec.CLD, inMsg: codec.ALI, expNextState: codec.CLD, expOutMsg: nil, expSend: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn := newConnWithState(tc.startState)
			in := makePkg(tc.inMsg)

			res, send := conn.listen(in)

			// State-Übergang
			assert.Equal(t, tc.expNextState, conn.state, "next state mismatch")

			// sendResponse
			assert.Equal(t, tc.expSend, send, "sendResponse mismatch")

			// Out MsgCode (falls spezifiziert)
			if tc.expOutMsg != nil {
				assert.Equal(t, *tc.expOutMsg, res.MSgCode, "response MSgCode mismatch")
			} else {
				// Bei silent-Fällen prüfen wir nur, dass wir NICHT gesendet haben
				assert.False(t, send, "expected silent (no send)")
			}

			// Kopierte Felder müssen immer gleich sein
			require.Equal(t, in.SessionID, res.SessionID, "SessionID must be copied")
			require.Equal(t, in.PackedID, res.PackedID, "PackedID must be copied")
			require.Equal(t, in.PackedID, res.FrameBegin, "FrameBegin must equal input PackedID")
			require.Equal(t, in.FrameEnd, res.FrameEnd, "FrameEnd must be copied")
			require.Equal(t, in.PayloadLength, res.PayloadLength, "PayloadLength must be copied")
		})
	}
}

// End-to-End Happy Path als Sequenztest: REQ -> OPN -> ACK -> ALI -> FIN -> CLD
func TestFSM_HappyPath_Sequence(t *testing.T) {
	conn := newConnWithState(codec.REQ)

	// REQ + REQ -> OPN
	res, send := conn.listen(makePkg(codec.REQ))
	require.True(t, send)
	require.Equal(t, codec.OPN, res.MSgCode)
	require.Equal(t, codec.OPN, conn.state)

	// OPN + OPN -> ACK
	res, send = conn.listen(makePkg(codec.OPN))
	require.True(t, send)
	require.Equal(t, codec.ACK, res.MSgCode)
	require.Equal(t, codec.ACK, conn.state)

	// ACK + ALI -> ALI
	res, send = conn.listen(makePkg(codec.ALI))
	require.True(t, send)
	require.Equal(t, codec.ALI, res.MSgCode)
	require.Equal(t, codec.ALI, conn.state)

	// ALI + FIN -> FIN
	res, send = conn.listen(makePkg(codec.FIN))
	require.True(t, send)
	require.Equal(t, codec.FIN, res.MSgCode)
	require.Equal(t, codec.FIN, conn.state)

	// FIN + FIN -> CLD
	res, send = conn.listen(makePkg(codec.FIN))
	require.True(t, send)
	require.Equal(t, codec.CLD, res.MSgCode)
	require.Equal(t, codec.CLD, conn.state)

	// CLD + (anything) -> silent, stays CLD
	_, send = conn.listen(makePkg(codec.ALI))
	require.False(t, send)
	require.Equal(t, codec.CLD, conn.state)
}

// Utility: Pointer auf State für optionale Erwartung
func ptr(s codec.State) *codec.State { return &s }
