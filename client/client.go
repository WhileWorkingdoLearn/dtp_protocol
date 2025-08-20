package main

import (
	"fmt"
	"net"

	dtp "github.com/WhilecodingDoLearn/dtp/protocol/dtp"
	protocol "github.com/WhilecodingDoLearn/dtp/protocol/types"
)

func main() {
	serverAddr := net.UDPAddr{
		Port: 8080,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	p := protocol.Package{
		Sid: 1234,
		Msg: protocol.REQ,
		Pid: 0,
		Bid: 0,
		Lid: 3,
		Pyl: []byte("ABC"),
		Rma: &serverAddr,
	}

	data := dtp.Encode(p)
	_, err = conn.Write(data)
	if err != nil {
		panic(err)
	}

	fmt.Println("Message sent to server.")

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		p, err := dtp.Decode(buffer[:n])
		if err != nil {
			panic(err)
		}
		fmt.Printf("Received '%v' from %s\n", p, p.Rma)
	}
}
