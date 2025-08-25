package main

import (
	"fmt"
	"testing"

	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	"github.com/stretchr/testify/assert"
)

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
