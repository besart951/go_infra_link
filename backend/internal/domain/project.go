package domain

type Project struct {
	ID      string `gorm:"primaryKey;size:36"`
	Name    string `gorm:"size:255"`
	OwnerID string `gorm:"size:36;index"`
}

type ProjectMember struct {
	ID        string `gorm:"primaryKey;size:36"`
	ProjectID string `gorm:"size:36;index"`
	UserID    string `gorm:"size:36;index"`
	Role      string `gorm:"size:32"`
}
