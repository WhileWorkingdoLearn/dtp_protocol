package main

import (
	"fmt"
	"net"
	"time"

	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	udpsim "github.com/WhilecodingDoLearn/dtp/pkg/protocol/dev/sim"
)

type ConnectionHandler struct {
	state    codec.State
	buffer   []codec.Package
	dataSize int
}

func handle(p codec.Package, connHandler *ConnectionHandler) (codec.Package, bool) {
	res := codec.Package{SessionID: p.SessionID, UserID: p.UserID, PackedID: p.PackedID, FrameBegin: p.FrameBegin, FrameEnd: p.FrameEnd, PayloadLength: p.PayloadLength, Payload: []byte{}, Rma: nil}
	send := false
	switch p.MSgCode {
	case codec.REQ:
		if connHandler.state == codec.ERR {
			res.MSgCode = codec.ERR
			if p.PayloadLength < 1024 {
				connHandler.dataSize = p.PayloadLength
				res.MSgCode = codec.OPN
				connHandler.state = codec.OPN
			}
			return res, true
		}
		if connHandler.state == codec.REQ {
			if p.PayloadLength < 1024 {
				connHandler.dataSize = p.PayloadLength
				res.MSgCode = codec.OPN
				connHandler.state = codec.OPN
			} else {
				res.MSgCode = codec.ERR
				connHandler.state = codec.ERR
			}
			return res, true
		}
		if connHandler.state == codec.OPN {
			return res, false
		}
	case codec.ACK:
		if connHandler.state == codec.OPN {
			connHandler.buffer = make([]codec.Package, connHandler.dataSize, connHandler.dataSize)
			connHandler.state = codec.ALI
			res.MSgCode = codec.ALI
			return res, true
		}
	case codec.ALI:
		if connHandler.state == codec.ALI {
			if len(connHandler.buffer) > 0 {

			}
		}
	case codec.ERR:

	default:
		{
			res = codec.Package{SessionID: p.SessionID}
		}
	}

	return res, send
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
		connHandler := ConnectionHandler{}

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

			res, send := handle(p, &connHandler)

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
