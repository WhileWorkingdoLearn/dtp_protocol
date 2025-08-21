package main

import (
	"fmt"
	"net"
	"time"

	"github.com/WhilecodingDoLearn/dtp/pkg/protocol/codec"
	udpsim "github.com/WhilecodingDoLearn/dtp/pkg/protocol/dev/sim"
)

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
		buf := make([]byte, 1024)
		var sessionState codec.State = 0
		for {
			n, addr, err := server.ReadFromUDP(buf)
			if err != nil {
				return
			}
			fmt.Printf("Server empfangen von %s: %s\n", addr, string(buf[:n]))
			// Echo
			p, err := codec.Decode(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}

			var res codec.Package
			hasResponse := true
			switch p.Msg {
			case codec.REQ:
				if sessionState == codec.REQ {
					res = codec.Package{Sid: p.Sid, Uid: 0, Msg: codec.OPN, Pid: 0, Bid: 0, Lid: 0, Tol: 0, Pyl: []byte{}, Rma: nil}
					sessionState = codec.OPN
				}
			default:
				{
					res = codec.Package{}
				}
			}
			if hasResponse {
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
	msg := codec.Encode(codec.Package{Sid: 123, Uid: 222, Msg: codec.REQ, Pid: 0, Bid: 0, Lid: 3, Tol: 0, Pyl: []byte{}, Rma: nil})
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
