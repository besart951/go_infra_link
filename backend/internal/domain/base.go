package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID, err = uuid.NewV7()
	}
	return
}

type PaginationParams struct {
	Page   int
	Limit  int
	Search string
}

type PaginatedList[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
}

func CalculateTotalPages(total int64, limit int) int {
	if limit == 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
