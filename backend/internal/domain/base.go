package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Base struct {
	ID        uuid.UUID  `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
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
	Page   int
	Limit  int
	Search string
	// Advanced filtering
	Filters map[string]interface{}
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
