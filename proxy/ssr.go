package proxy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ssrProxy struct {
	*Base
	Password      string `json:"password"`
	Cipher        string `json:"cipher"`
	Protocol      string `json:"protocol"`
	ProtocolParam string `json:"protocol_param,omitempty"`
	Obfs          string `json:"obfs"`
	ObfsParam     string `json:"obfs_param,omitempty"`
}

func newSSRByLink(link string) (*ssrProxy, error) {
	originLink := link
	link = strings.TrimPrefix(link, "ssr://")
	link = strings.ReplaceAll(link, "–", "+")
	link = strings.ReplaceAll(link, "_", "/")

	decodeLink, err := base64Decode(link)
	if err != nil {
		return nil, err
	}

	linkInfo := strings.SplitN(decodeLink, "/?", 2)
	if len(linkInfo) != 2 {
		return nil, fmt.Errorf("parser: invalid ssr link length: %s", link)
	}

	ssrInfo := strings.Split(linkInfo[0], ":")
	if len(ssrInfo) < 6 {
		return nil, fmt.Errorf("parser: invalid ssr info length: %s", link)
	}

	server := ssrInfo[0]
	port, err := strconv.Atoi(ssrInfo[1])
	if err != nil {
		return nil, fmt.Errorf("parser: invalid ssr port: %s", link)
	}
	protocol := ssrInfo[2]
	cipher := ssrInfo[3]
	obfs := ssrInfo[4]
	password, err := base64Decode(ssrInfo[5])
	if err != nil {
		return nil, fmt.Errorf("parser: invalid ssr password: %s", link)
	}

	params, err := url.ParseQuery(ssrInfo[1])
	if err != nil {
		return nil, fmt.Errorf("parser: invalid ssr params: %s", link)
	}

	// protocol param
	protocolParam, err := base64Decode(params.Get("protoparam"))
	if err != nil {
		return nil, fmt.Errorf("parser: invalid ssr protoparam: %s", link)
	}

	// obfs param
	obfsParam, err := base64Decode(params.Get("obfsparam"))
	if err != nil {
		return nil, fmt.Errorf("parser: invalid ssr obfsparam: %s", link)
	}

	// https://github.com/Dreamacro/clash/issues/865
	if cipher == "none" {
		cipher = "dummy"
	}

	return &ssrProxy{
		Base: &Base{
			Server: server,
			Port:   port,
			Type:   SSR,
			Link:   originLink,
		},
		Password:      password,
		Cipher:        cipher,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
	}, nil
}

func (p *ssrProxy) ConfigMap() (map[string]interface{}, error) {
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
	return m, nil
}
