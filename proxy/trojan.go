package proxy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type trojanProxy struct {
	*Base
	Password       string   `json:"password"`
	ALPN           []string `json:"alpn,omitempty"`
	SNI            string   `json:"sni,omitempty"`
	SkipCertVerify bool     `json:"skip-cert-verify"`
	UDP            bool     `json:"udp"`
}

func newTrojanByLink(link string) (*trojanProxy, error) {
	/**
	trojan-go://
	    $(trojan-password)
	    @
	    trojan-host
	    :
	    port
	/?
	    sni=$(tls-sni.com)&
	    type=$(original|ws|h2|h2+ws)&
	        host=$(websocket-host.com)&
	        path=$(/websocket/path)&
	    encryption=$(ss;aes-256-gcm;ss-password)&
	    plugin=$(...)
	#$(descriptive-text)
	*/

	uri, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	password := uri.User.Username()
	password, _ = url.QueryUnescape(password)

	server := uri.Hostname()
	port, _ := strconv.Atoi(uri.Port())

	moreInfos := uri.Query()
	sni := moreInfos.Get("sni")
	sni, _ = url.QueryUnescape(sni)
	transformType := moreInfos.Get("type")
	transformType, _ = url.QueryUnescape(transformType)
	host := moreInfos.Get("host")
	host, _ = url.QueryUnescape(host)
	path := moreInfos.Get("path")
	path, _ = url.QueryUnescape(path)

	alpn := make([]string, 0)
	if transformType == "h2" {
		alpn = append(alpn, "h2")
	}

	if port == 0 {
		return nil, fmt.Errorf("invalid port")
	}

	return &trojanProxy{
		Base: &Base{
			Server: server,
			Port:   port,
			Type:   "trojan",
		},
		Password:       password,
		ALPN:           alpn,
		UDP:            true,
		SNI:            host,
		SkipCertVerify: true,
	}, nil
}

func (p *trojanProxy) ConfigMap() (map[string]interface{}, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
