package structure

import (
	"net"

	dtp "github.com/WhilecodingDoLearn/dtp/protocol/connection"
)

type DTPAddr struct {
	Port int
	Ip   net.IP
	Zone string
}

type DTPConnection struct {
	udpConn *net.UDPConn
	reader  dtp.DTPReader
	writer
}

func Listen(protocol string, dtpAddress DTPAddr) (*DTPConnection, error) {
	serverAddr := net.UDPAddr{
		Port: dtpAddress.Port,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.DialUDP(protocol, nil, &serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dtpConn := DTPConnection{}
	return &dtpConn, nil

}
