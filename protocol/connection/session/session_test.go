package dtp

import (
	"net"
	"sync"
	"testing"

	protocol "github.com/WhilecodingDoLearn/dtp/protocol/types"
	"github.com/stretchr/testify/assert"
)

type Frame struct {
	start int
	end   int
}

type PacketCache struct {
	cache []struct {
		Sid int
		Msg int
		Pid int
		Bid int
		Lid int
		Pyl []byte
		Rma *net.UDPAddr
	}
	received int
}

type Buffer struct {
	frames     map[Frame]*PacketCache
	buffer     []string
	bufferSize int
	mux        sync.Mutex
}

/*
Connection has already been established.
 Packters arrive over udp. That means there are no guarantees that the Packages arrive in order or if they arrive.
 Packages need to be ordered. To reduce retransimition of lost or missing Packets they are organized in Frames.

*/

func (b *Buffer) add(packet struct {
	Sid int
	Msg int
	Pid int
	Bid int
	Lid int
	Pyl []byte
	Rma *net.UDPAddr
}) {

	frame := Frame{start: packet.Sid, end: packet.Lid}
	packetCache, ok := b.frames[frame]
	if !ok {
		newCache := make([]struct {
			Sid int
			Msg int
			Pid int
			Bid int
			Lid int
			Pyl []byte
			Rma *net.UDPAddr
		}, frame.end-frame.start+1, frame.end-frame.start+1)
		newCache[packet.Pid] = packet
		b.frames[frame] = &PacketCache{cache: newCache, received: 1}
	}
	if ok {
		packetCache.cache[packet.Pid] = packet
		packetCache.received++
		b.frames[frame] = packetCache
	}
}

func (b *Buffer) isFull(frame Frame) bool {

	cache, ok := b.frames[frame]
	if !ok {
		return ok
	}
	if cache.received < len(cache.cache) {
		return false
	}

	isComplete := true
	for idx, p := range cache.cache {
		if p.Sid+idx != p.Pid {
			isComplete = false
			break
		}
	}

	return isComplete
}

func (b *Buffer) Write(reader protocol.PackageReader) {
	p := reader.Read()

	if p.Pid < p.Sid || p.Pid > p.Lid {

	}

	b.add(p)

}

func TestBuffer(t *testing.T) {
	b := Buffer{
		frames:     make(map[Frame]*PacketCache),
		buffer:     make([]string, 0),
		bufferSize: 1000,
	}

	packet := struct {
		Sid int
		Msg int
		Pid int
		Bid int
		Lid int
		Pyl []byte
		Rma *net.UDPAddr
	}{Sid: 123, Pid: 0, Bid: 0, Lid: 3, Pyl: []byte{'A'}}

	b.add(packet)
	testframe := Frame{start: 0, end: 3}
	p := b.frames[testframe]
	assert.Len(t, p.cache, 4)
	assert.Equal(t, p.received, 1)

	isFull := b.isFull(testframe)
	assert.False(t, isFull)
}
