package internal

import (
	"fmt"
	"net"

	"p2p-chat/domain"
)

func NewPeerConnection(host string, port uint16) domain.Connection {
	return &peerConnection{
		host: host,
		port: port,
		conn: nil,
	}
}

type peerConnection struct {
	host string
	port uint16
	conn net.Conn
}

var _ domain.Connection = &peerConnection{}

func (pc *peerConnection) Open() error {
	if pc.conn != nil {
		return nil
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", pc.host, pc.port))
	if err != nil {
		return err
	}

	pc.conn = conn

	return nil
}

func (pc peerConnection) GetAddr() (host string, port uint16) {
	return pc.host, pc.port
}

func (pc peerConnection) Send(payload string) error {
	_, err := fmt.Fprintf(pc.conn, payload+"\n")
	return err
}

func (pc peerConnection) Close() error {
	return pc.conn.Close()
}
