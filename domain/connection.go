package domain

type Connection interface {
	Open()
	GetIP() string
	Send(payload []byte)
	Close()
}
