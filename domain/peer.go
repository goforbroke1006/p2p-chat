package domain

type Peer struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}
