package dtp

import (
	"fmt"
	"io"

	"github.com/WhilecodingDoLearn/dtp/protocol/dtp"
	protocol "github.com/WhilecodingDoLearn/dtp/protocol/types"
)

var sessionHandler = NewSessionHandler()

func Listen(connection io.ReadWriteCloser) {
	// receiver: Read into a fixed-size buffer

	for {
		buf := make([]byte, 256)
		n, err := connection.Read(buf)
		if err == io.EOF {
			fmt.Println("connection closed")
			break
		}
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		p, err := dtp.Decode(buf[:n])
		if err != nil {
			panic(err)
		}

		handle(p)

	}
}

func handle(p protocol.Package) {
	var session *Session

	//1 . Check if session is existent, if not create a new one
	s, ok := sessionHandler.GetSession(p.Sid)
	if !ok && p.Msg == REQ {
		session = NewSession(p.Sid)
		session.state = OPN
	} else {
		session = s
	}

	switch session.state {
	case OPN:
	case ACK:
	default:
		{
		}

	}

}

// 2.

/*
	if _, ok := sessionHandler.GetSession(p.Sid); ok {
		if p.Msg == protocol.REQ {
			fmt.Println("ignore")
			return
		}

	} else {

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


	switch
*/
