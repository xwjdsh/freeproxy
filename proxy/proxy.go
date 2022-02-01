package proxy

import (
	"encoding/base64"
	"encoding/json"
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
	SS  Type = "ss"
	SSR Type = "ssr"
)

type Base struct {
	Name   string `json:"name" gorm:"-"`
	Type   Type   `json:"type"`
	Server string `json:"server" gorm:"uniqueIndex:idx_server_port"`
	Port   int    `json:"port" gorm:"uniqueIndex:idx_server_port"`
	Link   string `json:"-"`
	Source string `json:"-"`

	Country     string `json:"-"`
	CountryCode string `json:"-"`
	Delay       uint16 `json:"-"`
}

func (b *Base) GetBase() *Base {
	return b
}

func (b *Base) Restore(cm string) (Proxy, error) {
	var proxy Proxy
	switch b.Type {
	case SS:
		proxy = &Shadowsocks{Base: b}
	case SSR:
		proxy = &ShadowsocksR{Base: b}
	}

	if err := json.Unmarshal([]byte(cm), proxy); err != nil {
		return nil, err
	}

	return proxy, nil
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
