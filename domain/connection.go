package domain

type Connection interface {
	Open() error
	GetAddr() (host string, port uint16)
	Send(payload string) error
	Close() error
}

type ConnectionsStorage interface {
	GetPeers() []Peer
	GetConnection(host string, port uint16) (result Connection)
	OnNewRemoteConnection(host string, port uint16, conn Connection)
	Run()
	Shutdown()
}
