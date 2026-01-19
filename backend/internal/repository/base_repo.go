package repository

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"

	"gorm.io/gorm"
)

// Generic function to handle pagination & search
func paginate[T any](db *gorm.DB, params domain.PaginationParams, searchFields []string) (*domain.PaginatedList[T], error) {
	var items []T
	var total int64

	query := db.Model(new(T))

	// Global Search
	if params.Search != "" && len(searchFields) > 0 {
		searchQuery := db
		for i, field := range searchFields {
			if i == 0 {
				searchQuery = searchQuery.Where(field+" ILIKE ?", "%"+params.Search+"%")
			} else {
				searchQuery = searchQuery.Or(field+" ILIKE ?", "%"+params.Search+"%")
			}
		}
		query = query.Where(searchQuery)
	}

	// Count Total
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply Pagination
	offset := (params.Page - 1) * params.Limit
	err := query.Limit(params.Limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedList[T]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}
