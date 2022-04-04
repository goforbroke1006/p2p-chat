package main

import (
	"fmt"
	p2p_chat "github.com/goforbroke1006/p2p-chat"
	"github.com/pkg/errors"

	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: p2p-chat <username> <host> <port>")
		os.Exit(1)
	}
	portInt, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		panic(errors.Wrap(err, "wrong port format"))
	}
	var (
		username = os.Args[1]
		host     = os.Args[2]
		port     = uint16(portInt)
	)

	fmt.Printf("%s %s:%d\n", username, host, port)

	multicast := p2p_chat.NewMulticast(username, host, port)
	go func() {
		for {
			err := multicast.Listen(func(username string, host string, port uint16) {
				fmt.Println(host, port)
			})
			fmt.Println("ERROR", err)
			<-time.After(5 * time.Second)
		}
	}()
	go multicast.Send()

	<-time.After(60 * time.Second)
}
