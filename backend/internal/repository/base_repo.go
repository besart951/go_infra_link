package repository

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"

	"gorm.io/gorm"
)

// Generic function to handle pagination & search
func Paginate[T any](db *gorm.DB, params domain.PaginationParams, searchFields []string) (*domain.PaginatedList[T], error) {
	var items []T
	var total int64

	query := db.Model(new(T))

	// Global Search - use LIKE for SQLite compatibility, ILIKE for PostgreSQL
	if params.Search != "" && len(searchFields) > 0 {
		// Detect database type
		dialector := db.Dialector.Name()
		
		var searchQuery *gorm.DB
		for i, field := range searchFields {
			searchPattern := "%" + params.Search + "%"
			if dialector == "postgres" {
				if i == 0 {
					searchQuery = db.Where(field+" ILIKE ?", searchPattern)
				} else {
					searchQuery = searchQuery.Or(field+" ILIKE ?", searchPattern)
				}
			} else {
				// SQLite and others use LIKE (case-insensitive by default in SQLite)
				if i == 0 {
					searchQuery = db.Where(field+" LIKE ?", searchPattern)
				} else {
					searchQuery = searchQuery.Or(field+" LIKE ?", searchPattern)
				}
			}
		}
		if searchQuery != nil {
			query = query.Where(searchQuery)
		}
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
