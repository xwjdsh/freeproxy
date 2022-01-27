package proxy

type Proxy interface {
	GetProtocol() Protocol
	GetBase() *Base
	GetClashMapping() (map[string]interface{}, error)
}

var (
	_ Proxy = new(Shadowsocks)
)

type Protocol string

const (
	SS Protocol = "ss"
)

type Base struct {
	Name    string `json:"name"`
	Server  string `json:"server"`
	Port    int    `json:"port"`
	UDP     bool   `json:"udp"`
	Country string `json:"country"`
	Useable bool   `json:"useable"`
}

func (b *Base) GetBase() *Base {
	return b
}
