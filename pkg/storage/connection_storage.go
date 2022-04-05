package storage

import (
	"context"
	"fmt"
	"time"

	"p2p-chat/domain"
	"p2p-chat/pkg/pub_sub"
	ctime "p2p-chat/pkg/time"
)

type CreateConnectionFunc func(ctx context.Context, host string, port uint16) domain.Connection

// NewConnectionStorage create storage with initial size of cache
func NewConnectionStorage(
	createConnFn CreateConnectionFunc,
	initSize int,
	updatePeersListDuration time.Duration,
) *connectionStorage {
	return &connectionStorage{
		createConnFn:            createConnFn,
		updatePeersListDuration: updatePeersListDuration,
		cache:                   make(map[string]domain.Connection, initSize),
		readPeersChReq:          make(chan struct{}),
		readPeersChRes:          make(chan []domain.Peer),
		readConnChReq:           make(chan string),
		readConnChRes:           make(chan struct{}),
		readConnPS:              pub_sub.New(),
		writeConnChReq:          make(chan connState, 1024),
		writeConnChRes:          make(chan struct{}),
		stopInit:                make(chan struct{}),
		stopDone:                make(chan struct{}),
	}
}

type connectionStorage struct {
	createConnFn            CreateConnectionFunc
	updatePeersListDuration time.Duration

	cache map[string]domain.Connection

	readPeersChReq chan struct{}
	readPeersChRes chan []domain.Peer
	readConnChReq  chan string
	readConnChRes  chan struct{}
	readConnPS     pub_sub.Node
	writeConnChReq chan connState
	writeConnChRes chan struct{}

	stopInit chan struct{}
	stopDone chan struct{}
}

var _ domain.ConnectionsStorage = &connectionStorage{}

func (c connectionStorage) GetPeers() []domain.Peer {
	c.readPeersChReq <- struct{}{}
	peers := <-c.readPeersChRes
	return peers
}

// GetConnection try to get connection in 3 parallel ways
// 1. Expect for new remote connection
// 2. Expect reading existing connection
// 3. Await new opened connection (step skips if remote connection from step 1 appears)
func (c connectionStorage) GetConnection(host string, port uint16) (result domain.Connection) {

	address := c.key(host, port)

	notifyConnCh := make(chan interface{}, 3)
	c.readConnPS.Subscribe(address, notifyConnCh)
	c.reading(address)
	defer c.readConnPS.Unsubscribe(address, notifyConnCh)

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) { // try open new connection
		connection := c.createConnFn(ctx, host, port)
		err := connection.Open()
		if err != nil {
			// TODO: error log
			_ = err
			return
		}
		select {
		case <-ctx.Done():
			go func() { _ = connection.Close() }()
		default:
			c.writing(address, connection)
		}
	}(ctx)

	// wait for connection
	select {
	case chunk := <-notifyConnCh:
		cancel()
		result = chunk.(connState).conn
		break
	}

	return result
}

// OnNewRemoteConnection store new connection from remote peer
func (c connectionStorage) OnNewRemoteConnection(host string, port uint16, conn domain.Connection) {
	c.writing(c.key(host, port), conn)
}

// Run process next type of operations
// * stopping storage pipelines and clear it
// * reading from cache
// * writing to cache
func (c *connectionStorage) Run() {

	var peers []domain.Peer
	updatePeersTicker := ctime.NewTicker(c.updatePeersListDuration, true)

RunLoop:
	for {
		select {
		case <-c.stopInit:
			break RunLoop

		case <-c.readPeersChReq:
			c.readPeersChRes <- peers

		case <-updatePeersTicker.C:
			if len(c.cache) > 0 {
				peers = make([]domain.Peer, 0, len(c.cache))
				for addr := range c.cache {
					host, port := c.cache[addr].GetAddr()
					peers = append(peers, domain.Peer{
						Host: host,
						Port: port,
					})
				}
			} else {
				peers = nil
			}

		case chunk := <-c.writeConnChReq:
			if old, found := c.cache[chunk.addr]; found {
				go func() { _ = old.Close() }()
			}
			c.cache[chunk.addr] = chunk.conn

			c.readConnPS.Publish(chunk.addr, connState{
				addr: chunk.addr,
				conn: chunk.conn,
			})

			c.writeConnChRes <- struct{}{}

		case addr := <-c.readConnChReq:
			if conn, found := c.cache[addr]; found {
				c.readConnPS.Publish(addr, connState{
					addr: addr,
					conn: conn,
				})
			}
			c.readConnChRes <- struct{}{}

		}
	}

	c.closeAllConnections()

	c.stopDone <- struct{}{}
}

// Shutdown break loop inside connectionStorage.Run() method and return to the 'owner' goroutine
func (c connectionStorage) Shutdown() {
	c.stopInit <- struct{}{}
	<-c.stopDone
}

// reading func send request for reading
// You have to subscribe on connectionStorage.readConnPS(fmt.Sprint("%d", ip), yourChannel) to read data
func (c connectionStorage) reading(addr string) {
	c.readConnChReq <- addr
	<-c.readConnChRes
}

// writing func send request for writing connection to cache
func (c connectionStorage) writing(addr string, conn domain.Connection) {
	c.writeConnChReq <- connState{
		addr: addr,
		conn: conn,
	}
	<-c.writeConnChRes
}

// closeAllConnections delete all keeping connection.
// Should be run after connectionStorage.Run loop break to prevent race on connectionStorage.cache
func (c connectionStorage) closeAllConnections() {
	for ip, conn := range c.cache {
		go func() { _ = conn.Close() }()
		delete(c.cache, ip)
	}
}

func (c connectionStorage) key(host string, port uint16) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// connState transport data about connection's changes
type connState struct {
	addr string
	conn domain.Connection
}
