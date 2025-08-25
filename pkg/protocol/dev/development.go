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
	state            codec.State
	buffer           FrameBuffer
	dataSize         int
	sessionId        int
	lastReceived     time.Time
	packagesReceived int
}

func (conn *DTPConnection) listen(p codec.Package) (res codec.Package, sendResponse bool) {
	// Wir gehen davon aus, dass wir antworten, außer wir setzen es explizit auf false
	sendResponse = true

	conn.packagesReceived++
	res.SessionID = p.SessionID
	res.PackedID = p.PackedID
	res.FrameBegin = p.PackedID
	res.FrameEnd = p.FrameEnd
	res.PayloadLength = p.PayloadLength

	// Validierung schlägt fehl → keine Antwort
	if err := conn.validate(p); err != nil {
		return res, false
	}
	conn.lastReceived = time.Now()

	switch conn.state {

	case codec.REQ:
		switch p.MSgCode {
		case codec.REQ:
			res.MSgCode = codec.OPN
			conn.state = codec.OPN
			return res, true
		default:
			res.MSgCode = codec.ERR
			return res, true
		}
	case codec.OPN:
		switch p.MSgCode {
		case codec.OPN:
			res.MSgCode = codec.ACK
			conn.state = codec.ACK
			return res, true
		default:
			res.MSgCode = codec.ERR
			return res, true
		}
	case codec.ACK:
		switch p.MSgCode {
		case codec.ACK:
			res.MSgCode = codec.ALI
			conn.state = codec.ALI
			return res, true
		default:
			res.MSgCode = codec.ERR
			return res, true
		}
	case codec.ALI:
		switch p.MSgCode {
		case codec.ALI:
			res.MSgCode = codec.ACK
			return res, false
		default:
			res.MSgCode = codec.ERR
			return res, true
		}
	case codec.RTY:
		switch p.MSgCode {
		case codec.RTY:
			res.MSgCode = codec.ACK
			return res, true
		default:
			res.MSgCode = codec.ERR
			return res, true
		}

	case codec.CLD:
		switch p.MSgCode {
		case codec.CLD:
			res.MSgCode = codec.ACK
			return res, true
		default:
			res.MSgCode = codec.ERR
			return res, true
		}

	default:
		return res, false
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
