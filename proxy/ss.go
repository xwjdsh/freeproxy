package proxy

import "encoding/json"

type Shadowsocks struct {
	Base
	Password   string                 `json:"password"`
	Cipher     string                 `json:"cipher"`
	Plugin     string                 `json:"plugin"`
	PluginOpts map[string]interface{} `json:"plugin-opts"`
}

func (p *Shadowsocks) GetProtocol() Protocol {
	return SS
}

func (p *Shadowsocks) GetClashMapping() (map[string]interface{}, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	m["type"] = string(p.GetProtocol())
	m["port"] = int(m["port"].(float64))
	return m, nil
}
