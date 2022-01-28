package storage

import (
	"gorm.io/gorm"

	"github.com/xwjdsh/freeproxy/proxy"
)

type Proxy struct {
	gorm.Model
	proxy.Base
}
