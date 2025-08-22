package dtp

import (
	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
)

type Writer interface {
	Write(msg Message) error
	Close()
}

type Reader interface {
	Read(b []byte) (*Message, error)
	Done() bool
}

type DTPHandler struct {
	buffer []byte
	cache  map[Frame][]codec.Package
}

func (dtpH DTPHandler) Read(b []byte) (*Message, error) {
	p, err := codec.Decode(b)
	if err != nil {

	}
	frm := Frame{start: p.FrameBegin, end: p.FrameEnd}
	ps, ok := dtpH.cache[frm]
	if ok {
		ps[p.PackedID] = p
		return nil, nil
	}
	ps = make([]codec.Package, p.FrameEnd-p.FrameBegin+1, p.FrameEnd-p.FrameBegin+1)
	ps[p.PackedID] = p

	return nil, nil
}

func (dtpH DTPHandler) Done() bool {

	return false
}
