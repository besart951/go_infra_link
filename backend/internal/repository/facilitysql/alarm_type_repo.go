package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Unit ---

type unitRepo struct {
	*gormbase.BaseRepository[*domainFacility.Unit]
}

func NewUnitRepository(db *gorm.DB) domainFacility.UnitRepository {
	search := func(q *gorm.DB, s string) *gorm.DB {
		p := "%" + strings.ToLower(strings.TrimSpace(s)) + "%"
		return q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", p, p)
	}
	return &unitRepo{gormbase.NewBaseRepository[*domainFacility.Unit](db, search)}
}

func (r *unitRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Unit], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

// --- AlarmField ---

type alarmFieldRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmField]
}

func NewAlarmFieldRepository(db *gorm.DB) domainFacility.AlarmFieldRepository {
	search := func(q *gorm.DB, s string) *gorm.DB {
		p := "%" + strings.ToLower(strings.TrimSpace(s)) + "%"
		return q.Where("LOWER(label) LIKE ? OR LOWER(key) LIKE ?", p, p)
	}
	return &alarmFieldRepo{gormbase.NewBaseRepository[*domainFacility.AlarmField](db, search)}
}

func (r *alarmFieldRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmField], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

// --- AlarmType ---

type alarmTypeRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmType]
	db *gorm.DB
}

func NewAlarmTypeRepository(db *gorm.DB) domainFacility.AlarmTypeRepository {
	search := func(q *gorm.DB, s string) *gorm.DB {
		p := "%" + strings.ToLower(strings.TrimSpace(s)) + "%"
		return q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", p, p)
	}
	return &alarmTypeRepo{
		BaseRepository: gormbase.NewBaseRepository[*domainFacility.AlarmType](db, search),
		db:             db,
	}
}

func (r *alarmTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 20)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *alarmTypeRepo) GetWithFields(id uuid.UUID) (*domainFacility.AlarmType, error) {
	var at domainFacility.AlarmType
	err := r.db.
		Preload("Fields").
		Preload("Fields.AlarmField").
		Preload("Fields.DefaultUnit").
		Where("id = ?", id).
		First(&at).Error
	if err != nil {
		return nil, err
	}
	return &at, nil
}

func (r *alarmTypeRepo) ListWithFields(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.AlarmType{})
	if params.Search != "" {
		p := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", p, p)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.AlarmType
	if err := query.
		Preload("Fields").
		Preload("Fields.AlarmField").
		Preload("Fields.DefaultUnit").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.AlarmType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

// --- AlarmTypeField ---

type alarmTypeFieldRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmTypeField]
}

func NewAlarmTypeFieldRepository(db *gorm.DB) domainFacility.AlarmTypeFieldRepository {
	return &alarmTypeFieldRepo{gormbase.NewBaseRepository[*domainFacility.AlarmTypeField](db, nil)}
}

func (r *alarmTypeFieldRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmTypeField], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
