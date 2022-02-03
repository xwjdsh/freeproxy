package proxy

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Vmess struct {
	*Base
	UUID           string            `json:"uuid"`
	AlterID        int               `json:"alterId"`
	Cipher         string            `json:"cipher"`
	TLS            bool              `json:"tls,omitempty"`
	Network        string            `json:"network,omitempty"`
	HTTPOpts       HTTPOptions       `json:"http-opts,omitempty"`
	WSPath         string            `json:"ws-path,omitempty"`
	WSHeaders      map[string]string `json:"ws-headers,omitempty"`
	SkipCertVerify bool              `json:"skip-cert-verify,omitempty"`
	ServerName     string            `json:"servername,omitempty"`
}

type HTTPOptions struct {
	Method  string              `json:"method,omitempty"`
	Path    []string            `json:"path,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
}

func NewVmessByLink(link string) (*Vmess, error) {
	originLink := link
	link = strings.TrimPrefix(link, "vmess://")
	decodeStr, err := base64Decode(link)
	if err != nil {
		return nil, err
	}

	var resp struct {
		V    string `json:"v"`
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		Aid  string `json:"aid"`
		Scy  string `json:"scy"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Sni  string `json:"sni"`
	}
	if err := json.Unmarshal([]byte(decodeStr), &resp); err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(resp.Port)
	if err != nil {
		return nil, err
	}
	alterId, _ := strconv.Atoi(resp.Aid)
	tls := resp.TLS == "tls"
	wsHeaders := make(map[string]string)
	if resp.Host != "" {
		wsHeaders["HOST"] = resp.Host
	}

	if resp.Path == "" {
		resp.Path = "/"
	}

	return &Vmess{
		Base: &Base{
			Server: resp.Add,
			Port:   port,
			Type:   VMESS,
			Link:   originLink,
		},
		UUID:           resp.ID,
		AlterID:        alterId,
		Cipher:         "auto",
		TLS:            tls,
		Network:        resp.Net,
		HTTPOpts:       HTTPOptions{},
		WSPath:         resp.Path,
		WSHeaders:      wsHeaders,
		SkipCertVerify: true,
		ServerName:     resp.Host,
	}, nil
}

func (p *Vmess) ConfigMap() (map[string]interface{}, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	m["port"] = int(m["port"].(float64))
	m["alterId"] = int(m["alterId"].(float64))
	return m, nil
}
