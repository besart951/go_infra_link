package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Base struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

func (b *Base) InitForCreate(now time.Time) error {
	if b.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		b.ID = id
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	return nil
}

func (b *Base) TouchForUpdate(now time.Time) {
	b.UpdatedAt = now
}

type PaginationParams struct {
	Page    int
	Limit   int
	Search  string
	OrderBy string
	Order   string // "asc" or "desc"
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
