package proxy

import (
	"encoding/base64"
	"fmt"
)

type Proxy interface {
	GetBase() *Base
	ConfigMap() (map[string]interface{}, error)
}

var (
	_ Proxy = new(Shadowsocks)
)

type Type string

const (
	SS Type = "ss"
)

type Base struct {
	Name   string `json:"name"`
	Type   Type   `json:"type"`
	Server string `json:"server"`
	Port   int    `json:"port"`
	UDP    bool   `json:"udp"`

	Country      string `json:"-"`
	CountryCode  string `json:"-"`
	CountryEmoji string `json:"-"`
}

func (b *Base) GetBase() *Base {
	return b
}

func base64Decode(src string) (string, error) {
	if src == "" {
		return "", nil
	}

	for _, encoder := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		data, err := encoder.DecodeString(src)
		if err == nil {
			return string(data), err
		}
	}

	return "", fmt.Errorf("base64 decode error")
}
