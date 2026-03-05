package repository

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"k8s_api_server/internal/config"
	"k8s_api_server/internal/model/entity"
)

func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Auto migrate
	if err := db.AutoMigrate(
		&entity.ResourceHistory{},
		&entity.User{},
		&entity.UserPermission{},
		&entity.Cluster{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
