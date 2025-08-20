package main

import (
	"io"
	"net"

	"time"

	udp_sim "github.com/WhilecodingDoLearn/dtp/development"
	connection "github.com/WhilecodingDoLearn/dtp/protocol/connection"
	dtp "github.com/WhilecodingDoLearn/dtp/protocol/dtp"
	protocol "github.com/WhilecodingDoLearn/dtp/protocol/types"
)

func initConnection(conn io.WriteCloser) {
	addr, err := net.ResolveUDPAddr("", ":8080")
	if err != nil {
		panic(err)
	}
	sid, err := connection.GenerateSessionId(0, 30001)
	if err != nil {

	}
	for i := 1; i <= 10; i++ {
		msg := protocol.Package{
			Sid: sid,
			Msg: protocol.REQ,
			Pid: 0,
			Bid: 0,
			Lid: 10,
			Pyl: []byte(""),
			Rma: addr,
		}
		_, _ = conn.Write(dtp.Encode(msg))
		time.Sleep(100 * time.Millisecond)
	}
	conn.Close()

}

func send(conn io.WriteCloser) {
	addr, err := net.ResolveUDPAddr("0.0.0.0", ":8080")
	if err != nil {
		panic(err)
	}

	for i := 1; i <= 10; i++ {
		msg := protocol.Package{
			Sid: 1234,
			Msg: protocol.ACK,
			Pid: i,
			Bid: 0,
			Lid: 10,
			Pyl: []byte("Hello"),
			Rma: addr,
		}
		_, _ = conn.Write(dtp.Encode(msg))
		time.Sleep(100 * time.Millisecond)
	}
	conn.Close()
}

func main() {
	// Create a “connection” with 200–500ms uniform jitter,
	// 10% loss, 5% dup, 20% reorder (buf cap 3).
	conn := udp_sim.DialUnreliableUDP(
		udp_sim.WithMaxDelay(500*time.Millisecond),
		udp_sim.WithLoss(0.1),
		udp_sim.WithDuplication(0.05),
		udp_sim.WithReordering(0.2, 3),
		udp_sim.WithUniformJitter(),
	)

	// sender: write 10 messages
	go initConnection(conn)

	connection.Listen(conn)

}
