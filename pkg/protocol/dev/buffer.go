package main

import (
	"fmt"

	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
)

const BufferSize = 1024

type Buffer struct {
	frames   [BufferSize]byte
	size     int
	received int
}

type FrameBuffer interface {
	Read(b codec.Package) error
	Flush()
	Size() int
}

func (b *Buffer) Read(p codec.Package) error {

	if p.FrameBegin < 0 || p.FrameBegin >= BufferSize {
		return fmt.Errorf("frame begin index %d out of range [0:%d]", p.FrameBegin, BufferSize-1)
	}
	if p.FrameEnd < 0 || p.FrameEnd >= BufferSize {
		return fmt.Errorf("frame end index %d out of range [0:%d]", p.FrameEnd, BufferSize-1)
	}

	if p.FrameBegin > p.FrameEnd {
		return fmt.Errorf("invalid frame range: begin %d > end %d", p.FrameBegin, p.FrameEnd)
	}

	expected := p.FrameEnd - p.FrameBegin + 1
	if len(p.Payload) != expected {
		return fmt.Errorf("payload length mismatch: got %d bytes, expected %d", len(p.Payload), expected)
	}

	// 4. Daten kopieren
	copy(b.frames[p.FrameBegin:p.FrameEnd+1], p.Payload)
	b.received += p.PayloadLength
	return nil
}

func (b *Buffer) Flush() {
	b.frames = [1024]byte{}
	b.received = 0
	b.size = 0
}

func (b *Buffer) Size() int {
	return b.size
}

func NewBuffer() FrameBuffer {
	return &Buffer{frames: [1024]byte{}}
}
