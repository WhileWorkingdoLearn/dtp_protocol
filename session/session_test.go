package session

import (
	"sync"
	"testing"

	"github.com/WhilecodingDpLearn/dtp/protocol"
	"github.com/stretchr/testify/assert"
)

type Frame struct {
	start int
	end   int
}

type PacketCache struct {
	cache    []protocol.Package
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

func (b *Buffer) add(packet protocol.Package) {

	frame := Frame{start: packet.Sid, end: packet.Lid}
	packetCache, ok := b.frames[frame]
	if !ok {
		newCache := make([]protocol.Package, frame.end-frame.start+1, frame.end-frame.start+1)
		newCache[packet.PId] = packet
		b.frames[frame] = &PacketCache{cache: newCache, received: 1}
	}
	if ok {
		packetCache.cache[packet.PId] = packet
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
		if p.Sid+idx != p.PId {
			isComplete = false
			break
		}
	}

	return isComplete
}

func (b *Buffer) Write(reader protocol.PackageReader) {
	p := reader.Read()

	if p.PId < p.Sid || p.PId > p.Lid {

	}

	b.add(p)

}

func TestBuffer(t *testing.T) {
	b := Buffer{
		frames:     make(map[Frame]*PacketCache),
		buffer:     make([]string, 0),
		bufferSize: 1000,
	}

	packet := protocol.Package{Sid: 123, PId: 0, Bid: 0, Lid: 3, Pyl: []byte{'A'}}

	b.add(packet)
	testframe := Frame{start: 0, end: 3}
	p := b.frames[testframe]
	assert.Len(t, p.cache, 4)
	assert.Equal(t, p.received, 1)

	isFull := b.isFull(testframe)
	assert.False(t, isFull)
}
