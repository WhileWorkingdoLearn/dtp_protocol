package main

import (
	"fmt"
	"net"

	"github.com/WhilecodingDoLearn/dtp/protocol"
)

func main() {
	addr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("UDP server listening on port 8080...")

	sessions := map[int]bool{}

	for {

		buffer := make([]byte, 1024)

		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		data, err := protocol.Decode(buffer[:n])
		if err != nil {
			panic(err)
		}

		fmt.Printf("Received '%v' from %s\n", data, clientAddr)

		if _, ok := sessions[data.Sid]; !ok && data.Msg == protocol.REQ {
			fmt.Println("Wants to connect")
			nP := protocol.Package{
				Sid: 1234,
				Msg: protocol.OPN,
				Pid: 0,
				Bid: 0,
				Lid: 0,
				Pyl: []byte{},
				Rma: &addr,
			}
			d := protocol.Encode(nP)
			conn.WriteTo(d, clientAddr)
			fmt.Println("Send msg to")
			break
		}

	}

}
