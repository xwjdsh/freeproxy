package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Dreamacro/clash/adapter"
	"github.com/stretchr/testify/assert"
)

func TestUrlTest(t *testing.T) {
	config := `{"cipher":"aes-256-gcm","name":"test","password":"g5MeD6Ft3CWlJId","plugin":"","plugin-opts":{},"port":5004,"server":"167.88.62.62","type":"ss","udp":false}`

	m := map[string]interface{}{}
	assert.Nil(t, json.Unmarshal([]byte(config), &m))
	m["port"] = int(m["port"].(float64))

	clashProxy, err := adapter.ParseProxy(m)
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	delay, err := clashProxy.URLTest(ctx, "http://www.gstatic.com/generate_204")
	assert.Nil(t, err)

	fmt.Println(delay)
}
