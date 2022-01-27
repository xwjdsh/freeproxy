package storage

import (
	"gorm.io/gorm"

	"github.com/xwjdsh/proxypool/proxy"
)

type Proxy struct {
	gorm.Model
	proxy.Base
}
