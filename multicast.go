package p2p_chat

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type RemotePeerHandler func(username string, host string, port uint16)

type Multicast interface {
	Listen(handler RemotePeerHandler) error
	Send()
}

func NewMulticast(username string, host string, port uint16) Multicast {
	addr, err := net.ResolveUDPAddr("udp", "233.0.0.0:9999")
	if err != nil {
		panic(errors.Wrap(err, "can't resolve address of peer"))
	}
	return &multicast{
		multicastAddr: addr,
		username:      username,
		host:          host,
		port:          port,
	}
}

type multicast struct {
	multicastAddr *net.UDPAddr
	username      string
	host          string
	port          uint16
}

var _ Multicast = &multicast{}

func (m multicast) Listen(handler RemotePeerHandler) error {

	conn, err := net.ListenMulticastUDP("udp", nil, m.multicastAddr)
	if err != nil {
		return err
	}

	for {
		payload := make([]byte, 1024)
		n, src, err := conn.ReadFromUDP(payload)
		_ = src
		if err != nil {
			return err
		}
		payload = payload[:n]

		text := string(payload)
		parts := strings.Split(text, ":")
		if len(parts) != 4 {
			continue
			//return err
		}

		if parts[0] != "p2p-chat" {
			continue
			//return nil
		}

		portInt, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			//return err
			continue
		}

		port := uint16(portInt)
		_ = port

		//handler(src.IP.String(), port)
		//handler(parts[1], src.IP.String(), uint16(src.Port))
		handler(parts[1], parts[2], port)
	}

	return nil
}

func (m multicast) Send() {
	for {
		conn, err := net.DialUDP("udp", nil, m.multicastAddr)
		if err != nil {
			fmt.Println("ERROR", err)
			continue
		}
		for {
			_, err := conn.Write([]byte(fmt.Sprintf("p2p-chat:%s:%s:%d", m.username, m.host, m.port)))
			if err != nil {
				fmt.Println("ERROR", err)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
