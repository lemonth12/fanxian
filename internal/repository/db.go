package repository

import (
	"fanxian/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(path string, wal bool) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if wal {
		db.Exec("PRAGMA journal_mode=WAL")
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Order{}); err != nil {
		return nil, err
	}
	return db, nil
}
