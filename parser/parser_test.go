package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSSLink(t *testing.T) {
	ss, err := parseSSLink(`ss://YWVzLTI1Ni1nY206S2l4THZLendqZWtHMDBybQ@38.68.134.37:8080#%E7%BF%BB%E5%A2%99%E5%85%9Afanqiangdang.com%400123_US_8`)
	require.Nil(t, err)
	assert.Equal(t, ss.Base.Server, "38.68.134.37")
	assert.Equal(t, ss.Base.Port, 8080)
	assert.Equal(t, ss.Password, "KixLvKzwjekG00rm")
	assert.Equal(t, ss.Cipher, "aes-256-gcm")
	assert.Empty(t, ss.Plugin)
	assert.Empty(t, ss.PluginOpts)
}
