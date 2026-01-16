package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(driver string, dsn string) (*gorm.DB, error) {
	if driver != "sqlite" {
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
