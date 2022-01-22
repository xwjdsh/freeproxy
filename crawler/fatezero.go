package crawler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type fatezero struct {
}

var fatezeroInstance = &fatezero{}

func (fatezero) name() string {
	return "fatezero"
}

func (c fatezero) fetch(ch chan *Proxy) {
	resp, err := http.Get("http://proxylist.fatezero.org/proxy.list")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result struct {
		Anonymity     string   `json:"anonymity"`
		Port          int      `json:"port"`
		Country       string   `json:"country"`
		ResponseTime  float64  `json:"response_time"`
		From          string   `json:"from"`
		ExportAddress []string `json:"export_address"`
		Type          string   `json:"type"`
		Host          string   `json:"host"`
	}
	for _, line := range strings.Split(string(data), "\n") {
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		proxy := &Proxy{
			Protocol: result.Type,
			Source:   c.name(),
			Host:     result.Host,
			Port:     result.Port,
		}
		switch result.Anonymity {
		case "transparent":
			proxy.Type = Transparent
		case "high_anonymous":
			proxy.Type = HighAnonymous
		}

		ch <- proxy
	}
}
