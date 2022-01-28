package storage

import (
	"context"
	"encoding/json"
	"fmt"
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
	}
}

type Handler struct {
	validateChan <-chan *proxy.Proxy
	db           *gorm.DB
}

func Init(cfg *config.StorageConfig) (*Handler, error) {
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.Driver {
	case "sqlite":
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

func (h *Handler) Store(ctx context.Context) {
	for {
		select {
		case p := <-h.validateChan:
			if p == nil {
				return
			}
			pp, err := NewProxy(p)
			if err != nil {

			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *Handler) SaveProxy(ctx context.Context, p proxy.Proxy) error {
	pp := &Proxy{
		Base:              p.GetBase(),
		LastAvailableTime: time.Now(),
	}
	h.db.WithContext(ctx).Create(p)
}
