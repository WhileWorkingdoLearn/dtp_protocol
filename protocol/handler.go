package protocol

import (
	"fmt"
	"io"

	"github.com/WhilecodingDoLearn/dtp/session"
)

func Handle(connection io.ReadWriteCloser) {
	// receiver: Read into a fixed-size buffer
	sessionHandler := session.NewSessionHandler()

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
		p, err := Decode(buf[:n])
		if err != nil {
			panic(err)
		}

		if _, ok := sessionHandler.GetSession(p.Sid); ok {
			if p.Msg == REQ {
				fmt.Println("ignore")
				continue
			}

		} else {

			if !ok && p.Msg == REQ && p.Pid == 0 {
				fmt.Println("Init session")
				session, err := sessionHandler.NewSession(p.Sid)
				if err != nil {
					panic(err)
				}
				session.ChangeState(OPN)
				sessionHandler.AddSession(session)

			}
			fmt.Printf("got packet: %v\n", p)
		}

	}
}
