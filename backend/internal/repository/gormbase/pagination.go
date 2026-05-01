package gormbase

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"gorm.io/gorm"
)

type OffsetPage struct {
	Page   int
	Limit  int
	Offset int
}

func NormalizeOffsetPage(params domain.PaginationParams, defaultLimit int) OffsetPage {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, defaultLimit)
	return OffsetPage{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func ExactOffsetPage[T any](query *gorm.DB, page OffsetPage, order any) (*domain.PaginatedList[T], error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var items []T
	if err := query.Session(&gorm.Session{}).
		Order(order).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[T]{
		Items:      items,
		Total:      total,
		Page:       page.Page,
		TotalPages: domain.CalculateTotalPages(total, page.Limit),
	}, nil
}
