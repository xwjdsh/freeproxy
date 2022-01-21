package crawler

import (
	"testing"
)

func TestDaili66(t *testing.T) {
	daili66Instance.page = 1
	result, err := daili66Instance.Fetch()
	if err != nil {
		t.Error(err)
	}
	if len(result.Proxies) == 0 {
		t.Error("empty result")
	}
	if result.Proxies[0].Host == "" ||
		result.Proxies[0].Port == 0 ||
		result.Proxies[0].Location == "" {
		t.Errorf("unexpected proxy: %+v", result.Proxies[0].Location)
	}
}
