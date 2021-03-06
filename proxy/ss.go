package proxy

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ssProxy struct {
	*Base
	Password   string                 `json:"password"`
	Cipher     string                 `json:"cipher"`
	Plugin     string                 `json:"plugin"`
	PluginOpts map[string]interface{} `json:"plugin-opts"`
}

type ssProxyHandler struct{}

func newSSByLink(link string) (*ssProxy, error) {
	uri, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	cipher := ""
	password := ""
	if uri.User.String() == "" {
		infos, err := base64Decode(uri.Hostname())
		if err != nil {
			return nil, err
		}
		uri, err = url.Parse("ss://" + infos)
		if err != nil {
			return nil, err
		}
		cipher = uri.User.Username()
		password, _ = uri.User.Password()
	} else {
		cipherInfoString, err := base64Decode(uri.User.Username())
		if err != nil {
			return nil, fmt.Errorf("proxy: [ss] %w", err)
		}
		cipherInfo := strings.SplitN(cipherInfoString, ":", 2)
		if len(cipherInfo) < 2 {
			return nil, fmt.Errorf("proxy: [ss] password parse error")
		}
		cipher = strings.ToLower(cipherInfo[0])
		password = cipherInfo[1]
	}
	server := uri.Hostname()
	port, _ := strconv.Atoi(uri.Port())

	moreInfos := uri.Query()
	pluginString := moreInfos.Get("plugin")
	plugin := ""
	pluginOpts := make(map[string]interface{})
	if strings.Contains(pluginString, ";") {
		pluginInfos, err := url.ParseQuery(pluginString)
		if err == nil {
			if strings.Contains(pluginString, "obfs") {
				plugin = "obfs"
				pluginOpts["mode"] = pluginInfos.Get("obfs")
				pluginOpts["host"] = pluginInfos.Get("obfs-host")
			} else if strings.Contains(pluginString, "v2ray") {
				plugin = "v2ray-plugin"
				pluginOpts["mode"] = pluginInfos.Get("mode")
				pluginOpts["host"] = pluginInfos.Get("host")
				pluginOpts["tls"] = strings.Contains(pluginString, "tls")
			}
		}
	}
	if port == 0 || cipher == "" {
		return nil, fmt.Errorf("proxy: [ss] invalid link")
	}

	return &ssProxy{
		Base: &Base{
			Type:   SS,
			Server: server,
			Port:   port,
			Link:   link,
		},
		Password:   password,
		Cipher:     cipher,
		Plugin:     plugin,
		PluginOpts: pluginOpts,
	}, nil
}

func (p *ssProxy) ConfigMap() (map[string]interface{}, error) {
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
