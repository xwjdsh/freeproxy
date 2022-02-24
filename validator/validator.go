package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Dreamacro/clash/adapter"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Validator struct {
	cfg *config.ValidatorConfig
}

func New(cfg *config.ValidatorConfig) *Validator {
	return &Validator{
		cfg: cfg,
	}
}

func (v *Validator) CheckNetwork(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.cfg.TestNetworkURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	return nil
}

func (v *Validator) Validate(ctx context.Context, p proxy.Proxy) error {
	m, err := p.ConfigMap()
	if err != nil {
		return err
	}
	clashProxy, err := adapter.ParseProxy(m)
	if err != nil {
		return err
	}

	base := p.GetBase()
	base.Delay = 0
	for i := 0; i < v.cfg.TestURLCount; i++ {
		base.Delay, err = func() (uint16, error) {
			ctx, cancel := context.WithTimeout(ctx, v.cfg.TestURLTimeout)
			defer cancel()

			return clashProxy.URLTest(ctx, v.cfg.TestURL)
		}()
		if err != nil {
			return err
		}
	}

	if base.Delay == 0 {
		return fmt.Errorf("invalid delay")
	}

	return nil
}

func (v *Validator) GetCountryInfo(ctx context.Context, server string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, v.cfg.GetCountryInfoTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://ip-api.com/json/%s?fields=countryCode,country", server), nil)
	if err != nil {
		return "", "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("validator: GetCountryInfo unexpected status code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var res struct {
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return "", "", err
	}

	return res.CountryCode, res.Country, nil
}

func (c *Validator) GetTestURL() string {
	return c.cfg.TestURL
}
