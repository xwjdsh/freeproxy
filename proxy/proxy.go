package proxy

type Proxy interface {
	GetProtocol() Protocol
	GetBase() *Base
}

var (
	_ Proxy = new(Shadowsocks)
)

type Protocol string

const (
	SS Protocol = "ss"
)

type Base struct {
	Server  string `json:"server,omitempty"`
	Port    int    `json:"port,omitempty"`
	UDP     bool   `json:"udp,omitempty"`
	Country string `json:"country,omitempty"`
	Useable bool   `json:"useable,omitempty"`
}

func (b *Base) GetBase() *Base {
	return b
}

type Shadowsocks struct {
	Base       `json:"-"`
	Password   string                 `json:"password,omitempty"`
	Cipher     string                 `json:"cipher,omitempty"`
	Plugin     string                 `json:"plugin,omitempty"`
	PluginOpts map[string]interface{} `json:"plugin-opts,omitempty"`
}

func (p *Shadowsocks) GetProtocol() Protocol {
	return SS
}
