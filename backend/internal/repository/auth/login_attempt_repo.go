package auth

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type loginAttemptRepo struct {
	db *gorm.DB
}

func NewLoginAttemptRepository(db *gorm.DB) domainAuth.LoginAttemptRepository {
	return &loginAttemptRepo{db: db}
}

func (r *loginAttemptRepo) Create(attempt *domainAuth.LoginAttempt) error {
	now := time.Now().UTC()
	if attempt.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		attempt.ID = id
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = now
	}
	return r.db.Create(attempt).Error
}

func (r *loginAttemptRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainAuth.LoginAttempt], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.Model(&domainAuth.LoginAttempt{})
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(email) LIKE ? OR LOWER(ip) LIKE ? OR LOWER(user_agent) LIKE ? OR LOWER(failure_reason) LIKE ?", pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainAuth.LoginAttempt
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainAuth.LoginAttempt]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
