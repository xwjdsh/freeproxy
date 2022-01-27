package proxy

type Proxy interface {
	GetProtocol() Protocol
	GetBase() *Base
	GetClashMapping() (map[string]interface{}, error)
	UpdateCountryInfo(string, string, string)
}

var (
	_ Proxy = new(Shadowsocks)
)

type Protocol string

const (
	SS Protocol = "ss"
)

type Base struct {
	Name         string `json:"name"`
	Server       string `json:"server"`
	Port         int    `json:"port"`
	UDP          bool   `json:"udp"`
	Country      string `json:"-"`
	CountryCode  string `json:"-"`
	CountryEmoji string `json:"-"`
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) UpdateCountryInfo(countryCode, country, countryEmoji string) {
	b.CountryCode, b.Country, b.CountryEmoji = countryCode, country, countryEmoji
}
