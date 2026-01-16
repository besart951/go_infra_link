package domain

type Object struct {
	ID        string `gorm:"primaryKey;size:36"`
	ProjectID string `gorm:"size:36;index"`
	Name      string `gorm:"size:255"`
}

type ObjectPermission struct {
	ID       string `gorm:"primaryKey;size:36"`
	ObjectID string `gorm:"size:36;index"`
	UserID   string `gorm:"size:36;index"`
	Role     string `gorm:"size:32"`
}
