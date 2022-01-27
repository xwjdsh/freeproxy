package storage

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/xwjdsh/proxypool/config"
)

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
