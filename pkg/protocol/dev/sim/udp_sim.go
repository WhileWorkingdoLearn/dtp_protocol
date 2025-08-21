package udpsim

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

// SimConfig steuert Paketverlust, Verzögerung und Reordering.
type SimConfig struct {
	LossRate    float64       // Wahrscheinlicher Paketverlust [0.0..1.0]
	MinDelay    time.Duration // Minimale Verzögerung pro Paket
	MaxDelay    time.Duration // Maximale Verzögerung pro Paket
	ReorderRate float64       // Chance, zusätzliche Verzögerung zu applizieren (Reordering)
}

// Config enthält die globale Simulationseinstellung (kann zur Laufzeit geändert werden).
var Config = SimConfig{
	LossRate:    0.0,
	MinDelay:    0,
	MaxDelay:    0,
	ReorderRate: 0.0,
}

func init() {
	// Initialisiere Zufallsgenerator für Delay/Reorder
	rand.Seed(time.Now().UnixNano())
}

// UDPAddr entspricht net.UDPAddr
type UDPAddr struct {
	IP   net.IP
	Port int
}

func (a *UDPAddr) Network() string { return "udp" }
func (a *UDPAddr) String() string  { return fmt.Sprintf("%s:%d", a.IP.String(), a.Port) }

// intern verwaltete Registry von Ports zu Kanälen
var (
	registry   = make(map[int]chan packet)
	registryMu sync.Mutex
)

// packet definiert die UDP-Paket-Nachricht
type packet struct {
	data []byte
	addr *UDPAddr
}

// UDPConn simuliert net.UDPConn
type UDPConn struct {
	local         *UDPAddr
	remote        *UDPAddr
	inbox         chan packet
	closed        chan struct{}
	closeOnce     sync.Once
	readDeadline  time.Time
	writeDeadline time.Time
}

// ListenUDP öffnet eine simulierte "bind"-Verbindung
func ListenUDP(laddr *UDPAddr) (*UDPConn, error) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[laddr.Port]; exists {
		return nil, errors.New("port bereits belegt")
	}

	inbox := make(chan packet, 1024)
	registry[laddr.Port] = inbox

	return &UDPConn{
		local:  laddr,
		inbox:  inbox,
		closed: make(chan struct{}),
	}, nil
}

// DialUDP verbindet lokal zu remote
func DialUDP(laddr, raddr *UDPAddr) (*UDPConn, error) {
	lc, err := ListenUDP(laddr)
	if err != nil {
		return nil, err
	}
	lc.remote = raddr
	return lc, nil
}

// ReadFromUDP liest ein Paket und liefert Absenderadresse
func (c *UDPConn) ReadFromUDP(b []byte) (int, *UDPAddr, error) {
	var timer <-chan time.Time
	if !c.readDeadline.IsZero() {
		dur := time.Until(c.readDeadline)
		if dur <= 0 {
			return 0, nil, errors.New("read deadline exceeded")
		}
		timer = time.After(dur)
	}

	select {
	case <-c.closed:
		return 0, nil, errors.New("connection closed")
	case pkt := <-c.inbox:
		n := copy(b, pkt.data)
		return n, pkt.addr, nil
	case <-timer:
		return 0, nil, errors.New("read deadline exceeded")
	}
}

// WriteToUDP schreibt ein Paket an addr (mit Loss, Delay, Reorder)
func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error) {
	registryMu.Lock()
	ch, exists := registry[addr.Port]
	registryMu.Unlock()
	if !exists {
		return 0, errors.New("Zielport nicht erreichbar")
	}

	// 1) Paketverlust
	if rand.Float64() < Config.LossRate {
		return len(b), nil // Paket geht verloren, aber wir tun so, als sei es gesendet
	}

	// 2) Basis-Verzögerung zufällig zwischen MinDelay und MaxDelay
	var delay time.Duration
	if Config.MaxDelay > Config.MinDelay {
		delta := Config.MaxDelay - Config.MinDelay
		delay = Config.MinDelay + time.Duration(rand.Int63n(int64(delta)))
	}

	// 3) Reordering: gelegentlich noch mehr Verzögerung draufpacken
	if rand.Float64() < Config.ReorderRate && Config.MaxDelay > 0 {
		extra := time.Duration(rand.Int63n(int64(Config.MaxDelay)))
		delay += extra
	}

	// 4) Deferred Send in eigener Goroutine
	pkt := packet{data: append([]byte(nil), b...), addr: c.local}
	go func() {
		select {
		case <-c.closed:
			return
		case <-time.After(delay):
			select {
			case ch <- pkt:
			case <-c.closed:
			}
		}
	}()

	return len(b), nil
}

// Read und Write nutzen ReadFromUDP/WriteToUDP mit gespeicherter remote-Adresse
func (c *UDPConn) Read(b []byte) (int, error) {
	n, _, err := c.ReadFromUDP(b)
	return n, err
}

func (c *UDPConn) Write(b []byte) (int, error) {
	return c.WriteToUDP(b, c.remote)
}

// Close schließt die Verbindung
func (c *UDPConn) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
		registryMu.Lock()
		delete(registry, c.local.Port)
		registryMu.Unlock()
	})
	return nil
}

// LocalAddr und RemoteAddr
func (c *UDPConn) LocalAddr() net.Addr  { return c.local }
func (c *UDPConn) RemoteAddr() net.Addr { return c.remote }

// Deadlines setzen
func (c *UDPConn) SetDeadline(t time.Time) error {
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

func (c *UDPConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *UDPConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}
