package facilitysql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Unit ---

type unitRepo struct {
	*gormbase.BaseRepository[*domainFacility.Unit]
}

func NewUnitRepository(db *gorm.DB) domainFacility.UnitRepository {
	return &unitRepo{gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.Unit](searchspec.Units.SearchColumns("")...),
	)}
}

func (r *unitRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Unit], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 50)
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
	return &alarmFieldRepo{gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.AlarmField](searchspec.AlarmFields.SearchColumns("")...),
	)}
}

func (r *alarmFieldRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmField], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 50)
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
	return &alarmTypeRepo{
		BaseRepository: gormbase.NewBaseRepository(db, alarmTypeSearchCallback()),
		db:             db,
	}
}

func (r *alarmTypeRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 20)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *alarmTypeRepo) GetWithFields(ctx context.Context, id uuid.UUID) (*domainFacility.AlarmType, error) {
	var at domainFacility.AlarmType
	err := r.db.WithContext(ctx).
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

func (r *alarmTypeRepo) ListWithFields(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmType], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.AlarmType{})
	if params.Search != "" {
		query = alarmTypeSearchCallback()(query, params.Search)
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

func alarmTypeSearchCallback() gormbase.SearchCallback[*domainFacility.AlarmType] {
	return gormbase.TrigramSearchCallback[*domainFacility.AlarmType](searchspec.AlarmTypes.SearchColumns("")...)
}

// --- AlarmTypeField ---

type alarmTypeFieldRepo struct {
	*gormbase.BaseRepository[*domainFacility.AlarmTypeField]
}

func NewAlarmTypeFieldRepository(db *gorm.DB) domainFacility.AlarmTypeFieldRepository {
	return &alarmTypeFieldRepo{gormbase.NewBaseRepository[*domainFacility.AlarmTypeField](db, nil)}
}

func (r *alarmTypeFieldRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.AlarmTypeField], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 50)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
