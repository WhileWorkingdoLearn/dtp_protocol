package main

import (
	"fmt"
	"net"
	"time"

	dtp "github.com/WhilecodingDoLearn/dtp/pkg/protocol"
	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	udpsim "github.com/WhilecodingDoLearn/dtp/pkg/protocol/dev/sim"
)

type DTPConnection struct {
	prevState        codec.State
	state            codec.State
	buffer           FrameBuffer
	dataSize         int
	sessionId        int
	lastReceived     time.Time
	lastChangeId     int
	lastChangeTime   time.Time
	packagesReceived int
}

// listen verarbeitet den FSM-Flow ausschließlich über State + Guard-Helper.
//
// Happy Path (logisch):
// REQ -> OPN -> ACK -> ALI -> FIN -> CLD
//
// Nebenpfade:
// - ALI: needRetry -> RTY | shouldFinish -> FIN | shouldCloseNow -> CLD | isAlive -> keepalive (silent)
// - RTY: retryOK -> ALI | else -> ERR
// - FIN: wantReopen -> OPN | else -> CLD
// - ERR: wantsOpen -> OPN | else -> ERR (mit Antwort)
//
// Beachte die Priorität in ALI: Retry > Finish > CloseNow > Keepalive > sonst ERR.

func (conn *DTPConnection) listen(p codec.Package) (res codec.Package, sendResponse bool) {
	sendResponse = false
	conn.packagesReceived++
	res.SessionID = p.SessionID
	res.PackedID = p.PackedID
	res.FrameBegin = p.PackedID
	res.FrameEnd = p.FrameEnd
	res.PayloadLength = p.PayloadLength

	current, next := chanceConnState(conn.state, p)
	if current != next {
		conn.lastChangeId = p.PackedID
		conn.lastChangeTime = time.Now()
	}
	conn.prevState = current
	conn.state = next
	res.MSgCode = next
	if current == codec.ALI && next == codec.ALI {

		return res, false
	}
	if current == codec.CLD && next == codec.CLD {
		return res, false
	}
	return res, true
}

// --- Guards: müssen je nach Protokoll implementiert werden ---
func wantsConnect(p codec.Package) bool {
	return p.MSgCode == codec.REQ
}
func wantsOpen(p codec.Package) bool      { return p.MSgCode == codec.OPN }
func isAlive(p codec.Package) bool        { return p.MSgCode == codec.ALI }
func needRetry(p codec.Package) bool      { return p.MSgCode == codec.RTY }
func shouldFinish(p codec.Package) bool   { return p.MSgCode == codec.FIN }
func shouldCloseNow(p codec.Package) bool { return p.MSgCode == codec.CLD }
func retryOK(p codec.Package) bool        { return p.MSgCode == codec.ACK }
func wantReopen(p codec.Package) bool     { return p.MSgCode == codec.OPN }

func chanceConnState(connState codec.State, p codec.Package) (prevState codec.State, nextState codec.State) {
	switch connState {

	case codec.REQ:
		// Start: Gegenstück signalisiert Verbindungswunsch?
		if wantsConnect(p) {
			return connState, codec.OPN
		}
		return connState, connState

	case codec.OPN:
		// Öffnen/Handshake-Schritt fortsetzen?
		if wantsOpen(p) {
			return connState, codec.ACK
		}
		if shouldCloseNow(p) {
			return connState, codec.CLD
		}
		return connState, codec.ERR

	case codec.ACK:
		// Nächster Schritt nach ACK ist "alive/ready" -> ALI
		if isAlive(p) {
			return connState, codec.ALI
		}
		return connState, codec.ALI

	case codec.ALI:
		// Priorisierte Abzweigungen aus der Arbeitsphase
		if needRetry(p) {
			return connState, codec.RTY
		}
		if shouldFinish(p) {
			return connState, codec.FIN
		}
		if shouldCloseNow(p) {
			return connState, codec.CLD
		}
		if isAlive(p) {
			// Keepalive: nichts senden
			return connState, connState
		}
		// Unerwartetes in ALI -> Fehler
		return connState, codec.ERR

	case codec.RTY:
		// Retry-Phase: entweder zurück nach ALI oder Fehler
		if retryOK(p) {
			return connState, codec.ALI
		}
		return connState, codec.ERR

	case codec.FIN:
		// Abschluss: Reopen oder endgültig schließen
		if wantReopen(p) {
			return connState, codec.OPN
		}
		// Default: schließen
		return connState, codec.CLD

	case codec.CLD:
		// Terminal: keine Antworten mehr
		return connState, codec.CLD

	case codec.ERR:
		// Recovery nur via "wantsOpen" (Neustart/Handshake)
		if wantsOpen(p) {
			return connState, codec.OPN
		}
		// in ERR verbleiben
		return connState, codec.ERR

	default:
		// Unbekannter State -> Fehlerzustand
		return connState, codec.ERR
	}

}

func (connHandler *DTPConnection) Write(p codec.Package) {

}

func (connHandler *DTPConnection) validate(p codec.Package) error {
	//SessionID
	if p.SessionID < 0 {
		return &dtp.PacketError{
			Text:     "wrong sessions id",
			Want:     0,
			Has:      p.SessionID,
			PacketID: p.PackedID,
		}
	}

	//SessionID
	if p.MSgCode != codec.REQ && p.SessionID != connHandler.sessionId {
		return &dtp.PacketError{
			Text:     "wrong sessions id",
			Want:     connHandler.sessionId,
			Has:      p.SessionID,
			PacketID: p.PackedID,
		}
	}

	//MSgCode
	if p.MSgCode != connHandler.state && p.MSgCode != codec.RTY && p.MSgCode != codec.ERR {
		return &dtp.PacketError{
			Text:     "illigal packet state",
			Want:     int(connHandler.state),
			Has:      int(p.MSgCode),
			PacketID: p.PackedID,
		}
	}
	// PackedID
	if p.PackedID < 0 {
		return &dtp.PacketError{
			Text:     "illigal packet id",
			Want:     0,
			Has:      p.PackedID,
			PacketID: p.PackedID,
		}
	}
	//FrameBegin
	if p.FrameBegin < 0 || p.FrameBegin > BufferSize {
		return &dtp.PacketError{
			Text:     "frame begin out of range",
			Want:     0,
			Has:      p.FrameBegin,
			PacketID: p.PackedID,
		}
	}
	//FrameEnd
	if p.FrameEnd < 0 || p.FrameEnd > BufferSize {
		return &dtp.PacketError{
			Text:     "frame end out of range",
			Want:     BufferSize,
			Has:      p.FrameEnd,
			PacketID: p.PackedID,
		}
	}
	// PayloadLength
	if p.PayloadLength < 0 || p.PayloadLength > BufferSize {
		return &dtp.PacketError{
			Text:     "invalid payload size",
			Want:     BufferSize,
			Has:      p.PayloadLength,
			PacketID: p.PackedID,
		}
	}
	// Payload
	if len(p.Payload) != p.PayloadLength {
		return &dtp.PacketError{
			Text:     "corrupt payload size",
			Want:     p.PayloadLength,
			Has:      len(p.Payload),
			PacketID: p.PackedID,
		}
	}
	//Rma * net.UDPAddr...
	return nil
}

func main() {
	// Simulation konfigurieren
	udpsim.Config.LossRate = 0.1 // 10% Pakete verworfen
	udpsim.Config.MinDelay = 10 * time.Millisecond
	udpsim.Config.MaxDelay = 100 * time.Millisecond
	udpsim.Config.ReorderRate = 0.2 // 20% zusätzliche Verzögerung

	// Server starten
	serverAddr := &udpsim.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999}
	server, err := udpsim.ListenUDP(serverAddr)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	go func() {
		connHandler := DTPConnection{}

		for {
			readBuf := make([]byte, 1024)
			n, addr, err := server.ReadFromUDP(readBuf)
			if err != nil {
				return
			}
			fmt.Printf("Server empfangen von %s: %s\n", addr, string(readBuf[:n]))
			// Echo
			p, err := codec.Decode(readBuf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}

			res, send := connHandler.listen(p)

			if send {
				data := codec.Encode(res)
				server.WriteToUDP(data, addr)
			}

		}
	}()

	// Client initialisieren
	clientAddr := &udpsim.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000}
	client, err := udpsim.DialUDP(clientAddr, serverAddr)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Nachricht senden und Antwort lesen
	msg := codec.Encode(codec.Package{SessionID: 123, UserID: 222, MSgCode: codec.REQ, PackedID: 0, FrameBegin: 0, FrameEnd: 3, PayloadLength: 0, Payload: []byte{}, Rma: nil})
	client.Write(msg)
	client.SetReadDeadline(time.Now().Add(10 * time.Second))

	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		fmt.Println("Fehler beim Lesen:", err)
		return
	}
	fmt.Printf("Client empfangen: %s\n", string(buf[:n]))
}
