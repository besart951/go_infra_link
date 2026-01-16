package domain

type User struct {
	ID    string `gorm:"primaryKey;size:36"`
	Email string `gorm:"uniqueIndex;size:255"`
}
