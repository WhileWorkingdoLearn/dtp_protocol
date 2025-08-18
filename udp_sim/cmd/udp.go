package main

import (
	"fmt"
	"io"
	"net"

	"time"

	"github.com/WhilecodingDpLearn/dtp/protocol"
	"github.com/WhilecodingDpLearn/dtp/session"
	"github.com/WhilecodingDpLearn/dtp/udp_sim"
)

func initConnection(conn *udp_sim.Conn) {
	addr, err := net.ResolveUDPAddr("", ":8080")
	if err != nil {
		panic(err)
	}
	sid, err := session.GenerateSessionId(0, 30001)
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
		_, _ = conn.Write(protocol.Encode(msg))
		time.Sleep(100 * time.Millisecond)
	}
	conn.Close()

}

func send(conn *udp_sim.Conn) {
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
		_, _ = conn.Write(protocol.Encode(msg))
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

	// receiver: Read into a fixed-size buffer
	sessionHandler := session.NewSessionHandler()

	for {
		buf := make([]byte, 256)
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Println("connection closed")
			break
		}
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		p, err := protocol.Decode(buf[:n])
		if err != nil {
			panic(err)
		}

		session, ok := sessionHandler.GetSession(p.Sid)
		if ok {
			fmt.Println("has session")
			if p.Pid == 0 {
				fmt.Println("ignore")
				continue
			}
			if session.IsOpen() {
				fmt.Println("is Open")
			}

		}

		if !ok && p.Msg == protocol.REQ && p.Pid == 0 {
			fmt.Println("Init session")
			session, err := sessionHandler.NewSession(p.Sid)
			if err != nil {
				panic(err)
			}
			session.ChangeState(protocol.OPN)
			sessionHandler.AddSession(session)
		}
		fmt.Printf("got packet: %v\n", p)
	}

}
