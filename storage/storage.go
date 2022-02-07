package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/xwjdsh/freeproxy/config"
	"github.com/xwjdsh/freeproxy/proxy"
)

type Proxy struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	*proxy.Base
	Config string
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
		Base:   p.GetBase(),
		Config: string(data),
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
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
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
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{Logger: newLogger})
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

func (h *Handler) Remove(ctx context.Context, id uint) error {
	return h.db.Delete(&Proxy{}, id).Error
}

func (h *Handler) Update(ctx context.Context, p *Proxy) error {
	b := p.GetBase()
	return h.db.Model(p).Updates(map[string]interface{}{"delay": b.Delay, "country_code": b.CountryCode, "country": b.Country}).Error
}

func (h *Handler) Create(ctx context.Context, p proxy.Proxy) (*Proxy, bool, error) {
	m, err := p.ConfigMap()
	if err != nil {
		return nil, false, err
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, false, err
	}

	pp := &Proxy{
		Base:   p.GetBase(),
		Config: string(data),
	}

	r := h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "server"}, {Name: "port"}},
		DoNothing: true,
	}).Create(pp)
	if r.Error != nil {
		return nil, false, r.Error
	}

	return pp, r.RowsAffected > 0, nil
}

func (h *Handler) GetProxies(ctx context.Context) ([]*Proxy, error) {
	proxies := []*Proxy{}
	return proxies, h.db.Find(&proxies).Error
}

func (h *Handler) GetProxy(ctx context.Context, id uint, countryCode string) (*Proxy, error) {
	p := new(Proxy)
	db := h.db.Order("RANDOM()")
	if id != 0 {
		db = db.Where("id = ?", id)
	}
	if countryCode != "" {
		db = db.Where("country_code = ?", countryCode)
	}
	return p, db.First(p).Error
}
