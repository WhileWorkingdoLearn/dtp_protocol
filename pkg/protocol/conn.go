package dtp

import "net"

type Conn interface {
	ReadMessage() (*Message, error)
	WriteMessage(*Message) error
	Close() error
}

type DTPConnection struct {
	conn           *net.UDPConn
	dtpReader      Reader
	sessionHandler SessionHandler
}

func NewDTP() (Conn, error) {

	return &DTPConnection{}, nil
}

func (c *DTPConnection) ReadMessage() (*Message, error) {

	var msg *Message
	for !c.dtpReader.Done() {
		buffer := make([]byte, 1024)
		n, err := c.conn.Read(buffer)
		if err != nil {
			return nil, err
		}

		msg, err = c.dtpReader.Read(buffer[:n])
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (c DTPConnection) WriteMessage(*Message) error {
	return nil
}

func (dtpC DTPConnection) Close() error {
	return nil
}
