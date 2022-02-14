package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

var ErrInvalidLink = fmt.Errorf("proxy: invalid link")

type Proxy interface {
	GetBase() *Base
	ConfigMap() (map[string]interface{}, error)
}

var (
	_ Proxy = new(ssProxy)
)

type Type = string

const (
	SS    Type = "ss"
	SSR   Type = "ssr"
	Vmess Type = "vmess"
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
		proxy = &ssProxy{Base: b}
	case SSR:
		proxy = &ssrProxy{Base: b}
	case Vmess:
		proxy = &vmessProxy{Base: b}
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

func NewProxyByLink(link string) (Proxy, error) {
	var (
		p   Proxy
		err error
	)
	switch {
	case strings.HasPrefix(link, "ss://"):
		p, err = newSSByLink(link)
	case strings.HasPrefix(link, "ssr://"):
		p, err = newSSRByLink(link)
	case strings.HasPrefix(link, "vmess://"):
		p, err = newVmessByLink(link)
	default:
		err = ErrInvalidLink
	}

	return p, err
}

func LinkValid(link string) bool {
	return strings.HasPrefix(link, "ss://") ||
		strings.HasPrefix(link, "ssr://") ||
		strings.HasPrefix(link, "vmess://")
}