package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Proxy struct {
	gorm.Model
	*proxy.Base
	LastAvailableTime time.Time
	Config            string
}

func NewProxy(p proxy.Proxy) (*Proxy, error) {
	m, err := p.ConfigMap()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		Base:              p.GetBase(),
		LastAvailableTime: time.Now(),
		Config:            string(data),
	}, nil
}

type Result struct {
	Proxy *Proxy
	Error error
}

type Handler struct {
	db *gorm.DB
}

func Init(cfg *config.StorageConfig) (*Handler, error) {
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.Driver {
	case "sqlite":
		if _, err := os.Stat(cfg.DSN); os.IsNotExist(err) {
			_ = os.MkdirAll(path.Dir(cfg.DSN), 0755)
			f, err := os.Create(cfg.DSN)
			if err != nil {
				return nil, err
			}
			f.Close()
		}
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	}
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Proxy{}); err != nil {
		return nil, fmt.Errorf("storage: db.AutoMigrate error: %w", err)
	}

	return &Handler{
		db: db,
	}, nil
}

func (h *Handler) Store(ctx context.Context, p proxy.Proxy) (*Proxy, error) {
	m, err := p.ConfigMap()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	pp := &Proxy{
		Base:              p.GetBase(),
		LastAvailableTime: time.Now(),
		Config:            string(data),
	}

	if err := h.db.WithContext(ctx).Create(pp).Error; err != nil {
		return nil, err
	}
	return pp, nil
}

func (h *Handler) GetProxies(ctx context.Context) ([]*Proxy, error) {
	proxies := []*Proxy{}
	return proxies, h.db.Find(&proxies).Error
}
