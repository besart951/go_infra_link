package domain

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&BusinessDetails{},
		&Project{},
		&Phase{},
		&Building{},
		&ControlCabinet{},
	)
}
