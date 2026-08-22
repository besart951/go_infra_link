package facilitysql

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type spsControllerRepo struct {
	*gormbase.BaseRepository[*domainFacility.SPSController]
	db *gorm.DB
}

func NewSPSControllerRepository(db *gorm.DB) domainFacility.SPSControllerRepository {
	baseRepo := gormbase.NewBaseRepository(db, spsControllerSearchCallback())
	return &spsControllerRepo{BaseRepository: baseRepo, db: db}
}

func (r *spsControllerRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	query := activeSPSControllers(r.db.WithContext(ctx).Model(&domainFacility.SPSController{}))
	if strings.TrimSpace(params.Search) != "" {
		query = spsControllerSearchCallback()(query, params.Search)
	}
	result, err := gormbase.ExactOffsetPage[*domainFacility.SPSController](
		query, gormbase.NormalizeOffsetPage(params, 10), clause.OrderByColumn{
			Column: clause.Column{Table: clause.CurrentTable, Name: "created_at"}, Desc: true,
		},
	)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *spsControllerRepo) GetPaginatedListByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID)
	query = activeSPSControllers(query)

	if strings.TrimSpace(params.Search) != "" {
		query = spsControllerSearchCallback()(query, params.Search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.SPSController
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSController]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *spsControllerRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	if len(ids) == 0 {
		return []*domainFacility.SPSController{}, nil
	}
	var items []*domainFacility.SPSController
	err := activeSPSControllers(r.db.WithContext(ctx).Model(&domainFacility.SPSController{})).
		Where("sps_controllers.id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *spsControllerRepo) Update(ctx context.Context, entity *domainFacility.SPSController) error {
	mutation := activeMutation{
		target: activeMutationRequest{table: "sps_controllers", id: entity.ID, query: func(tx *gorm.DB) *gorm.DB {
			return activeSPSControllers(tx.Model(&domainFacility.SPSController{}))
		}},
		mutate: func(tx *gorm.DB) error {
			repo := gormbase.NewBaseRepository[*domainFacility.SPSController](tx, spsControllerSearchCallback())
			return repo.Update(ctx, entity)
		},
	}
	return runActiveMutation(ctx, r.db, mutation)
}

func spsControllerSearchCallback() gormbase.SearchCallback[*domainFacility.SPSController] {
	return gormbase.TrigramSearchCallback[*domainFacility.SPSController](searchspec.SPSControllers.SearchColumns("")...)
}

func (r *spsControllerRepo) GetIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *spsControllerRepo) GetIDsByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(controlCabinetIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id IN ?", controlCabinetIDs).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *spsControllerRepo) ListGADevicesByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]string, error) {
	var devices []string
	err := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID).
		Where("ga_device IS NOT NULL").
		Where("TRIM(ga_device) <> ''").
		Pluck("ga_device", &devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *spsControllerRepo) ExistsDeviceName(ctx context.Context, controlCabinetID uuid.UUID, deviceName string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID).
		Where("LOWER(device_name) = ?", strings.ToLower(strings.TrimSpace(deviceName)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *spsControllerRepo) ExistsGADevice(ctx context.Context, controlCabinetID uuid.UUID, gaDevice string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID).
		Where("UPPER(ga_device) = ?", strings.ToUpper(strings.TrimSpace(gaDevice)))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *spsControllerRepo) ExistsIPAddressVlan(ctx context.Context, ipAddress string, vlan string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&domainFacility.SPSController{}).
		Where("ip_address = ?", strings.TrimSpace(ipAddress)).
		Where("vlan = ?", strings.TrimSpace(vlan))

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *spsControllerRepo) GetByIdsForExport(ctx context.Context, ids []uuid.UUID) ([]domainFacility.SPSController, error) {
	if len(ids) == 0 {
		return []domainFacility.SPSController{}, nil
	}
	var items []domainFacility.SPSController
	err := r.db.WithContext(ctx).
		Where("sps_controllers.id IN ?", ids).
		Preload("ControlCabinet").
		Preload("ControlCabinet.Building").
		Preload("SPSControllerSystemTypes").
		Find(&items).Error
	return items, err
}

func (r *spsControllerRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&domainFacility.SPSController{}).Error
}
