package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"p2p-chat/domain"
	"p2p-chat/pkg/pub_sub"
)

func Test_connectionStorage_GetPeers(t *testing.T) {
	const updateDuration = 100 * time.Millisecond

	type fields struct {
		updatePeersListDuration time.Duration
		cache                   map[string]domain.Connection
	}
	tests := []struct {
		name   string
		fields fields
		want   []domain.Peer
	}{
		{
			name: "empty",
			fields: fields{
				updatePeersListDuration: updateDuration,
				cache:                   map[string]domain.Connection{},
			},
			want: nil,
		},
		{
			name: "non empty",
			fields: fields{
				updatePeersListDuration: updateDuration,
				cache: map[string]domain.Connection{
					"192.168.2.45:8888": createFakeConnFn(context.Background(), "192.168.2.45", 8888),
					"192.168.2.47:9999": createFakeConnFn(context.Background(), "192.168.2.47", 9999),
				},
			},
			want: []domain.Peer{
				{Host: "192.168.2.45", Port: 8888},
				{Host: "192.168.2.47", Port: 9999},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cs := connectionStorage{
				updatePeersListDuration: tt.fields.updatePeersListDuration,
				cache:                   tt.fields.cache,
				readPeersChReq:          make(chan struct{}),
				readPeersChRes:          make(chan []domain.Peer),
			}
			go cs.Run()
			<-time.After(time.Second)

			done := make(chan struct{})
			var peers []domain.Peer
			go func() {
				peers = cs.GetPeers()
				done <- struct{}{}
			}()

			select {
			case <-time.After(5 * time.Second):
				t.Fatal("should return data")
			case <-done:
				// ok
			}
			assert.Equalf(t, tt.want, peers, "GetPeers()")
		})
	}
}

func TestConnectionStorageWithPS_GetConnection(t *testing.T) {
	//t.Parallel()

	const size = 1024

	t.Run("connection can be established in 6 seconds", func(t *testing.T) {
		const (
			host = "192.168.1.19"
			port = uint16(12345)
		)

		t.Parallel()

		cs := NewConnectionStorage(createFakeConnFn, size, 15*time.Second)
		go cs.Run()
		defer cs.Shutdown()

		done := make(chan struct{})
		go func() {
			_ = cs.GetConnection(host, port)
			done <- struct{}{}
		}()

		select {
		case <-time.After(6 * time.Second):
			t.Error(errors.New("can't open connection in 6 seconds"))
		case <-done:
			// ok
		}
	})

	t.Run("same connection should be open once", func(t *testing.T) {
		t.Parallel()

		const (
			host = "192.168.1.23"
			port = uint16(23456)
		)

		cs := &connectionStorage{
			cache: make(map[string]domain.Connection, size),

			readConnChReq:  make(chan string),
			readConnChRes:  make(chan struct{}),
			readConnPS:     pub_sub.New(),
			writeConnChReq: make(chan connState, 1024),
			writeConnChRes: make(chan struct{}),

			stopInit: make(chan struct{}),
			stopDone: make(chan struct{}),
		}
		go cs.Run()
		defer cs.Shutdown()

		var (
			conn1 domain.Connection = nil
			conn2 domain.Connection = nil
			done                    = make(chan struct{})
		)

		go func() {
			conn1 = cs.GetConnection(host, port)
			done <- struct{}{}
		}()
		select {
		case <-time.After(6 * time.Second):
			t.Error(errors.New("can't open connection in 6 seconds"))
		case <-done:
			// ok
		}

		go func() {
			conn2 = cs.GetConnection(host, port)
			done <- struct{}{}
		}()
		select {
		case <-time.After(10 * time.Millisecond):
			t.Fatal(errors.New("can't open same connection immediately"))
		case <-done:
			// ok
		}

		if conn1 == nil {
			t.Fatal("should not be NIL")
		}

		if conn1 != conn2 {
			t.Fatalf("storage return different connections for same ip: %p and %p", conn1, conn2)
		}
	})

	t.Run("OnNewRemoteConnection + GetConnection(immediately)", func(t *testing.T) {
		t.Parallel()

		const (
			host = "192.168.1.33"
			port = uint16(23456)
		)

		var (
			remoteConn   domain.Connection = nil
			fromStorConn domain.Connection = nil
			done                           = make(chan struct{})
		)

		cs := NewConnectionStorage(createFakeConnFn, size, 15*time.Second)
		go cs.Run()
		defer cs.Shutdown()

		remoteConn = createFakeConnFn(context.Background(), host, port)
		cs.OnNewRemoteConnection(host, port, remoteConn)

		go func() {
			fromStorConn = cs.GetConnection(host, port)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}

		if remoteConn != fromStorConn {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn, fromStorConn)
		}
	})

	t.Run("GetConnection + OnNewRemoteConnection (immediately return from GetConnection)", func(t *testing.T) {
		t.Parallel()

		const (
			host = "192.168.1.46"
			port = uint16(12387)
		)

		var (
			remoteConn   domain.Connection = nil
			fromStorConn domain.Connection = nil
			start                          = make(chan struct{})
			done                           = make(chan struct{})
		)

		cs := NewConnectionStorage(createFakeConnFn, size, 15*time.Second)
		go cs.Run()
		defer cs.Shutdown()

		remoteConn = createFakeConnFn(context.Background(), host, port)

		go func() {
			start <- struct{}{}
			fromStorConn = cs.GetConnection(host, port)
			done <- struct{}{}
		}()

		<-start
		cs.OnNewRemoteConnection(host, port, remoteConn)

		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}

		if fromStorConn == nil {
			t.Fatal(errors.New("should not be NIL"))
		}

		if remoteConn != fromStorConn {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn, fromStorConn)
		}
	})

	t.Run("OnNewRemoteConnection + GetConnection + OnNewRemoteConnection+Connection.Close() + GetConnection - old closed, new opened", func(t *testing.T) {
		t.Parallel()

		const (
			host = "192.168.1.51"
			port = uint16(14394)
		)

		var (
			remoteConn1   domain.Connection = nil
			fromStorConn1 domain.Connection = nil

			remoteConn2   domain.Connection = nil
			fromStorConn2 domain.Connection = nil

			done = make(chan struct{})
		)

		cs := NewConnectionStorage(createFakeConnFn, size, 15*time.Second)
		go cs.Run()
		defer cs.Shutdown()

		remoteConn1 = createFakeConnFn(context.Background(), host, port)
		remoteConn1.(*fakeConnection).opened = true
		cs.OnNewRemoteConnection(host, port, remoteConn1)
		go func() {
			fromStorConn1 = cs.GetConnection(host, port)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}
		if remoteConn1 != fromStorConn1 {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn1, fromStorConn1)
		}

		remoteConn2 = createFakeConnFn(context.Background(), host, port)
		remoteConn2.(*fakeConnection).opened = true
		cs.OnNewRemoteConnection(host, port, remoteConn2)
		go func() {
			fromStorConn2 = cs.GetConnection(host, port)
			done <- struct{}{}
		}()
		select {
		case <-time.After(1 * time.Millisecond):
			t.Error(errors.New("can't get connection immediately"))
		case <-done:
			// ok
		}
		if remoteConn2 != fromStorConn2 {
			t.Errorf("storage return different connections for same ip: %p and %p", remoteConn2, fromStorConn2)
		}

		if fromStorConn1 == fromStorConn2 {
			t.Fatal("second connection should be new, not same as first connection")
		}

		<-time.After(1500 * time.Millisecond) // wait fake connection is closed
		assert.Equal(t, false, remoteConn1.(*fakeConnection).opened)
	})
}

// BenchmarkConnectionStorage_GetConnection checks efficiency open connection callback running
//
// go test -gcflags=-N -test.bench '^\QBenchmarkConnectionStorage_GetConnection\E$' -run ^$ -benchmem -test.benchtime 10000x ./...
func BenchmarkConnectionStorage_GetConnection(b *testing.B) {

	const size = 5
	const (
		host = "192.168.1.55"
		port = uint16(14394)
	)

	wrongConnCounter := uint64(0)

	for repeatIndex := 0; repeatIndex < b.N; repeatIndex++ {
		var (
			remoteConn      domain.Connection = nil
			fromStorageConn domain.Connection = nil
			start                             = make(chan struct{})
			done                              = make(chan struct{})
		)
		cs := NewConnectionStorage(createFakeConnFn, size, 15*time.Second)
		go cs.Run()

		go func() {
			start <- struct{}{}
			fromStorageConn = cs.GetConnection(host, port)
			done <- struct{}{}
		}()

		<-start
		remoteConn = createFakeConnFn(context.Background(), host, port)
		cs.OnNewRemoteConnection(host, port, remoteConn)

		<-done
		if remoteConn != fromStorageConn {
			wrongConnCounter++
		}
	}
	if wrongConnCounter > 0 {
		b.Errorf("for N = %d wrong connection return count %d", b.N, wrongConnCounter)
	}
}

// go test -gcflags=-N -test.bench '^\QBenchmarkConnectionStorage_Shutdown\E$' -run ^$ -benchmem -test.benchtime 10000x
func BenchmarkConnectionStorage_Shutdown(b *testing.B) {
	const startHostIndex = int32(101)
	const endHostIndex = int32(200)
	const port = uint16(14394)

	for repeatIndex := 0; repeatIndex < b.N; repeatIndex++ {
		cs := NewConnectionStorage(createFakeConnFn, int(endHostIndex-startHostIndex), 15*time.Second)
		go cs.Run()

		b.StopTimer()
		for hostIndex := startHostIndex; hostIndex <= endHostIndex; hostIndex++ {
			host := fmt.Sprintf("192.168.3.%d", hostIndex)
			conn := createFakeConnFn(context.Background(), host, port)
			cs.OnNewRemoteConnection(host, port, conn)
		}
		b.StartTimer()

		cs.Shutdown()
	}
}

var createFakeConnFn = func(ctx context.Context, host string, port uint16) domain.Connection {
	return &fakeConnection{
		host:   host,
		port:   port,
		opened: false,
	}
}

type fakeConnection struct {
	host   string
	port   uint16
	opened bool
}

var _ domain.Connection = &fakeConnection{}

func (fc fakeConnection) Open() error {
	time.Sleep(5 * time.Second)
	fc.opened = true
	return nil
}

func (fc fakeConnection) GetAddr() (host string, port uint16) {
	return fc.host, fc.port
}

func (fc fakeConnection) Send(_ string) error {
	return nil
}

func (fc fakeConnection) Close() error {
	time.Sleep(time.Second)
	fc.opened = false
	return nil
}
