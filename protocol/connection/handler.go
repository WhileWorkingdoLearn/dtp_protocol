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

		handle(p, sessionHandler)

	}
}

func Write(connection io.WriteCloser) {

}

func handle(p protocol.Package, sh *SessionHandler) {

	//1 . Check if session is existent, if not create a new one. Check if packags.Msg == Request new Connection

	switch p.Msg {
	case protocol.REQ:
		session, ok := sh.GetSession(p.Sid)
		if !ok {
			fmt.Println("Not found")
			session = NewSession(p.Sid)
			err := sh.AddSession(session)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		if session.state == protocol.REQ {
			fmt.Println("Change session state to OPN")
			session.state = protocol.OPN
		}

	case protocol.OPN:
		session, ok := sh.GetSession(p.Sid)
		if !ok {
			fmt.Println("error")
			return
		}
		if session.state == protocol.REQ {
			fmt.Println("Ignore")
			return
		}
		if session.state == protocol.OPN {
			fmt.Println("Set State to opwn")
			session.state = protocol.ACK
			return
		}

	case protocol.ACK:
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
