package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"p2p-chat/domain"
	"p2p-chat/internal"
	"p2p-chat/pkg/free_port"
	"p2p-chat/pkg/log"
	"p2p-chat/pkg/public_ip"
	"p2p-chat/pkg/storage"
)

func NewCliCmd() *cobra.Command {
	var (
		username string
		host            = ""
		port     uint16 = 0
	)
	cmd := &cobra.Command{
		Use: "cli",
		Run: func(cmd *cobra.Command, args []string) {

			if host == "" {
				publicIP, err := public_ip.Get()
				if err != nil {
					panic(err)
				}
				host = publicIP
			}

			if port == 0 {
				freePort, err := free_port.GetFreePort()
				if err != nil {
					panic(err)
				}
				port = uint16(freePort)
			}

			logger := log.NewLogger()
			logger.Infof("new node %s:%d starting...", host, port)
			logger.Infof("new node %s username", username)

			createConnFn := func(ctx context.Context, host string, port uint16) domain.Connection {
				logger.Infof("open new connection to %s %d", host, port)
				return internal.NewPeerConnection(host, port)
			}
			connectionStorage := storage.NewConnectionStorage(createConnFn, 1024, time.Minute)
			go connectionStorage.Run()
			defer connectionStorage.Shutdown()

			onRemotePeer := func(conn domain.Connection) {
				remoteHost, remotePort := conn.GetAddr()
				logger.Infof("new remote connection from %s %d", remoteHost, remotePort)
				connectionStorage.OnNewRemoteConnection(remoteHost, remotePort, conn)
			}
			tcpPeerHandler := internal.NewTcpPeerHandler(host, port, onRemotePeer)
			go tcpPeerHandler.RunListener()
			defer tcpPeerHandler.Shutdown()

			peers, err := internal.ReadPeerListFromFile("./peers.txt")
			if err != nil {
				logger.WithErr(err).Warn("can't read peers config")
			}
			logger.Infof("init with %d peers", len(peers))
			for _, peer := range peers {
				if peer.Host == host && peer.Port == port { // exclude current node
					continue
				}
				_ = connectionStorage.GetConnection(peer.Host, peer.Port)
			}

			// TODO: connect to peers from list

			<-time.After(10 * time.Minute)
		},
	}
	cmd.PersistentFlags().StringVarP(&username, "username", "u", username, "Chat username")
	cmd.PersistentFlags().StringVarP(&host, "host", "H", host, "Peer host")
	cmd.PersistentFlags().Uint16VarP(&port, "port", "P", port, "Peer port")
	return cmd
}
