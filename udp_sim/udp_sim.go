package udp_sim

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/WhilecodingDpLearn/dtp/protocol"
)

// JitterType controls the delay distribution.
type JitterType int

const (
	JitterUniform JitterType = iota
	JitterNormal
	JitterExponential
)

// Config holds all tunable parameters.
type Config struct {
	MaxDelay         time.Duration
	LossProb         float64
	DupProb          float64
	ReorderProb      float64
	ReorderBufferCap int
	Jitter           JitterType
	NormalStdDev     float64 // in nanoseconds
	ExpMean          float64 // in nanoseconds
	BandwidthLimit   int     // max in-flight packets
}

// Option is a functional option for Config.
type Option func(*Config)

// WithMaxDelay sets the upper bound for per‐packet latency.
func WithMaxDelay(d time.Duration) Option {
	return func(c *Config) { c.MaxDelay = d }
}

// WithLoss sets the probability [0,1) that a packet is dropped.
func WithLoss(p float64) Option {
	return func(c *Config) { c.LossProb = p }
}

// WithDuplication sets the chance [0,1) of sending a duplicate.
func WithDuplication(p float64) Option {
	return func(c *Config) { c.DupProb = p }
}

// WithReordering sets reorder probability and buffer capacity.
func WithReordering(p float64, bufCap int) Option {
	return func(c *Config) {
		c.ReorderProb = p
		c.ReorderBufferCap = bufCap
	}
}

// WithUniformJitter uses uniform [0,MaxDelay).
func WithUniformJitter() Option {
	return func(c *Config) { c.Jitter = JitterUniform }
}

// WithNormalJitter uses N(μ=MaxDelay/2, σ=stdDev).
func WithNormalJitter(stdDev time.Duration) Option {
	return func(c *Config) {
		c.Jitter = JitterNormal
		c.NormalStdDev = float64(stdDev)
	}
}

// WithExponentialJitter uses Exp(mean=meanDelay).
func WithExponentialJitter(meanDelay time.Duration) Option {
	return func(c *Config) {
		c.Jitter = JitterExponential
		c.ExpMean = float64(meanDelay)
	}
}

// WithBandwidthLimit caps the number of in-flight packets.
func WithBandwidthLimit(maxInFlight int) Option {
	return func(c *Config) { c.BandwidthLimit = maxInFlight }
}

// NewSimulator constructs your unreliable “link.”
func NewSimulator(opts ...Option) (
	sendCh chan<- protocol.Package,
	recvCh <-chan protocol.Package,
) {
	// Default parameters
	cfg := &Config{
		MaxDelay:         200 * time.Millisecond,
		LossProb:         0.0,
		DupProb:          0.0,
		ReorderProb:      0.0,
		ReorderBufferCap: 0,
		Jitter:           JitterUniform,
		NormalStdDev:     float64(50 * time.Millisecond),
		ExpMean:          float64(100 * time.Millisecond),
		BandwidthLimit:   0, // 0 = unlimited
	}
	for _, opt := range opts {
		opt(cfg)
	}

	rand.Seed(time.Now().UnixNano())

	send := make(chan protocol.Package)
	recv := make(chan protocol.Package)

	go func() {
		var (
			reorderBuf []protocol.Package
			inFlight   = make(chan struct{}, cfg.BandwidthLimit)
			mu         sync.Mutex
		)

		// helper to throttle in-flight packets
		acquire := func() {
			if cfg.BandwidthLimit > 0 {
				inFlight <- struct{}{}
			}
		}
		release := func() {
			if cfg.BandwidthLimit > 0 {
				<-inFlight
			}
		}

		// deliver a packet after randomized delay
		deliver := func(p protocol.Package) {
			go func(pkt protocol.Package) {
				delay := sampleDelay(cfg)
				time.Sleep(delay)
				recv <- pkt
				release()
			}(p)
		}

		// flush and shuffle reorder buffer
		flushReorder := func() {
			mu.Lock()
			if len(reorderBuf) == 0 {
				mu.Unlock()
				return
			}
			idx := rand.Perm(len(reorderBuf))
			for _, i := range idx {
				acquire()
				deliver(reorderBuf[i])
			}
			reorderBuf = reorderBuf[:0]
			mu.Unlock()
		}

		for pkt := range send {
			// Packet loss?
			if rand.Float64() < cfg.LossProb {
				continue
			}

			// How many copies?
			copies := 1
			if rand.Float64() < cfg.DupProb {
				copies = 2
			}

			for i := 0; i < copies; i++ {
				// Reordering?
				if rand.Float64() < cfg.ReorderProb && len(reorderBuf) < cfg.ReorderBufferCap {
					mu.Lock()
					reorderBuf = append(reorderBuf, pkt)
					mu.Unlock()
				} else {
					// push out any buffered packets first
					flushReorder()
					acquire()
					deliver(pkt)
				}
			}
		}

		// send remaining buffered packets
		flushReorder()
		close(recv)
	}()

	return send, recv
}

// sampleDelay returns a random delay based on cfg.Jitter.
func sampleDelay(cfg *Config) time.Duration {
	maxNs := float64(cfg.MaxDelay)
	switch cfg.Jitter {
	case JitterNormal:
		d := rand.NormFloat64()*cfg.NormalStdDev + maxNs/2
		if d < 0 {
			d = 0
		} else if d > maxNs {
			d = maxNs
		}
		return time.Duration(d)
	case JitterExponential:
		d := rand.ExpFloat64() * cfg.ExpMean
		if d > maxNs {
			d = maxNs
		}
		return time.Duration(d)
	default: // uniform
		return time.Duration(rand.Float64() * maxNs)
	}
}

// Conn wraps the unreliableudp simulator channels into
// a simple Read/Write interface.
type Conn struct {
	sendCh chan<- protocol.Package
	recvCh <-chan protocol.Package
	closed bool
}

// DialUnreliableUDP creates a new Conn with all the same
// knobs as unreliableudp.NewSimulator.
func DialUnreliableUDP(opts ...Option) *Conn {
	send, recv := NewSimulator(opts...)
	return &Conn{sendCh: send, recvCh: recv}
}

// Write sends the entire buffer as one packet.
func (c *Conn) Write(b []byte) (int, error) {
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	// Make a copy so caller can reuse b immediately
	data := make([]byte, len(b))
	copy(data, b)
	c.sendCh <- protocol.Package{Pyl: data}
	return len(b), nil
}

// Read blocks until the next packet arrives or the link closes.
// It copies up to len(p) bytes and returns the count.
func (c *Conn) Read(p []byte) (int, error) {
	pkt, ok := <-c.recvCh
	if !ok {
		return 0, io.EOF
	}
	n := copy(p, pkt.Pyl)
	return n, nil
}

// Close shuts down the sending side; any in-flight packets
// will still be delivered.
func (c *Conn) Close() error {
	if !c.closed {
		close(c.sendCh)
		c.closed = true
	}
	return nil
}
