package main

import (
	"testing"

	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestListen(t *testing.T) {
	type tc struct {
		name          string
		initState     codec.State
		inCode        codec.State
		payloadLength int

		wantState   codec.State
		wantSend    bool
		wantResCode codec.State
	}

	cases := []tc{
		{
			name:        "REQ→OPN (happy path)",
			initState:   codec.REQ,
			inCode:      codec.REQ,
			wantState:   codec.OPN,
			wantSend:    true,
			wantResCode: codec.OPN,
		},
		{
			name:        "REQ→invalid",
			initState:   codec.REQ,
			inCode:      codec.ACK,
			wantState:   codec.REQ,
			wantSend:    false,
			wantResCode: 0,
		},
		{
			name:        "OPN→ACK",
			initState:   codec.OPN,
			inCode:      codec.OPN,
			wantState:   codec.ACK,
			wantSend:    true,
			wantResCode: codec.ACK,
		},
		{
			name:        "OPN→invalid",
			initState:   codec.OPN,
			inCode:      codec.REQ,
			wantState:   codec.OPN,
			wantSend:    true,
			wantResCode: codec.ERR,
		},
		{
			name:        "ACK→ALI",
			initState:   codec.ACK,
			inCode:      codec.ACK,
			wantState:   codec.ALI,
			wantSend:    true,
			wantResCode: codec.ALI,
		},
		{
			name:        "ACK→invalid",
			initState:   codec.ACK,
			inCode:      codec.REQ,
			wantState:   codec.ACK,
			wantSend:    false,
			wantResCode: 0,
		},
		{
			name:        "ALI→DATA (no send)",
			initState:   codec.ALI,
			inCode:      codec.ALI,
			wantState:   codec.ALI,
			wantSend:    false,
			wantResCode: 0,
		},
		{
			name:        "ALI→RTY",
			initState:   codec.ALI,
			inCode:      codec.RTY,
			wantState:   codec.RTY,
			wantSend:    true,
			wantResCode: codec.ACK,
		},
		{
			name:        "ALI→CLD",
			initState:   codec.ALI,
			inCode:      codec.CLD,
			wantState:   codec.CLD,
			wantSend:    true,
			wantResCode: codec.ACK,
		},
		{
			name:        "ALI→invalid",
			initState:   codec.ALI,
			inCode:      codec.REQ,
			wantState:   codec.ERR,
			wantSend:    true,
			wantResCode: codec.ERR,
		},
		{
			name:          "RTY→small payload → ALI",
			initState:     codec.RTY,
			inCode:        codec.RTY,
			payloadLength: BufferSize - 1,
			wantState:     codec.ALI,
			wantSend:      true,
			wantResCode:   codec.ALI,
		},
		{
			name:          "RTY→too large → stay RTY",
			initState:     codec.RTY,
			inCode:        codec.RTY,
			payloadLength: BufferSize + 1,
			wantState:     codec.RTY,
			wantSend:      true,
			wantResCode:   codec.RTY,
		},
		{
			name:        "RTY→invalid",
			initState:   codec.RTY,
			inCode:      codec.ACK,
			wantState:   codec.ERR,
			wantSend:    true,
			wantResCode: codec.ERR,
		},
		{
			name:        "CLD→ACK → reset to REQ",
			initState:   codec.CLD,
			inCode:      codec.ACK,
			wantState:   codec.REQ,
			wantSend:    true,
			wantResCode: codec.CLD,
		},
		{
			name:        "CLD→invalid → still CLD",
			initState:   codec.CLD,
			inCode:      codec.OPN,
			wantState:   codec.CLD,
			wantSend:    true,
			wantResCode: codec.CLD,
		},
		{
			name:        "unknown state → ERR",
			initState:   codec.State(999),
			inCode:      codec.REQ,
			wantState:   codec.ERR,
			wantSend:    true,
			wantResCode: codec.ERR,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conn := &DTPConnection{
				state: c.initState,
			}

			// Input-Paket aufsetzen
			inPkg := codec.Package{
				MSgCode:       c.inCode,
				SessionID:     42,
				PackedID:      7,
				FrameEnd:      7,
				PayloadLength: c.payloadLength,
			}

			res, sent := conn.listen(inPkg)

			require.Equal(t, c.wantSend, sent, "sendResponse")
			require.Equal(t, c.wantResCode, res.MSgCode, "res.MSgCode")
			require.Equal(t, c.wantState, conn.state, "next state")
		})
	}
}
