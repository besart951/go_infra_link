package facilitysql

import (
	"strconv"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stateTextRepo struct {
	db *gorm.DB
}

func NewStateTextRepository(db *gorm.DB) domainFacility.StateTextRepository {
	return &stateTextRepo{db: db}
}

func (r *stateTextRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.StateText, error) {
	var items []*domainFacility.StateText
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *stateTextRepo) Create(entity *domainFacility.StateText) error {
	return r.db.Create(entity).Error
}

func (r *stateTextRepo) Update(entity *domainFacility.StateText) error {
	return r.db.Save(entity).Error
}

func (r *stateTextRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.StateText{}, ids).Error
}

func (r *stateTextRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.StateText], error) {
	var items []domainFacility.StateText
	var total int64

	db := r.db.Model(&domainFacility.StateText{})

	if params.Search != "" {
		// Try to parse search as int for RefNumber
		if refNum, err := strconv.Atoi(params.Search); err == nil {
			db = db.Where("ref_number = ?", refNum)
		} else {
			// Search in text fields if not a number, or just as a fallback/OR condition?
			// Since RefNumber is likely the primary search, let's stick to simple text search on the first text field
			db = db.Where("state_text1 ILIKE ?", "%"+params.Search+"%")
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("ref_number ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.StateText]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}
