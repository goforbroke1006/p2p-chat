package internal

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"p2p-chat/domain"
)

type TcpPeerHandler interface {
	RunListener()
	Shutdown()
}

type OnNewTcpConnectionCb func(conn domain.Connection)

func NewTcpPeerHandler(host string, port uint16, onRemotePeer OnNewTcpConnectionCb) *handler {
	return &handler{
		host:         host,
		port:         port,
		onRemotePeer: onRemotePeer,

		stopInit: make(chan struct{}),
		stopDone: make(chan struct{}),
	}
}

type handler struct {
	host         string
	port         uint16
	onRemotePeer OnNewTcpConnectionCb

	stopInit chan struct{}
	stopDone chan struct{}
}

func (h handler) RunListener() {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", h.host, h.port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Printf("Listening on %s:%d\n", h.host, h.port)

ListenLoop:
	for {
		select {
		case <-h.stopInit:
			break ListenLoop
		default:
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			// Handle connections in a new goroutine.
			remoteAddr := conn.RemoteAddr().String()
			remoteAddrParts := strings.Split(remoteAddr, ":")
			host := remoteAddrParts[0]
			port, _ := strconv.ParseInt(remoteAddrParts[1], 10, 64)

			connection := NewPeerConnection(host, uint16(port))

			go h.onRemotePeer(connection)
		}
	}
}

func (h handler) Shutdown() {
	h.stopInit <- struct{}{}
	<-h.stopDone
}
