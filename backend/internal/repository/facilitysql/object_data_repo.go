package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataRepo struct {
	*gormbase.BaseRepository[*domainFacility.ObjectData]
	db *gorm.DB
}

func (r *objectDataRepo) withObjectDataPreloads(query *gorm.DB) *gorm.DB {
	return query.
		Preload("BacnetObjects").
		Preload("BacnetObjects.AlarmType").
		Preload("Apparats")
}

func (r *objectDataRepo) getPaginatedListFiltered(projectID *uuid.UUID, apparatID *uuid.UUID, systemPartID *uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ObjectData{})
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}

	if apparatID != nil {
		sub := r.db.Table("object_data_apparats").
			Select("object_data_id").
			Where("apparat_id = ?", *apparatID)
		query = query.Where("id IN (?)", sub)
	}

	if systemPartID != nil {
		sub := r.db.Table("object_data_apparats AS oda").
			Select("DISTINCT oda.object_data_id").
			Joins("JOIN system_part_apparats spa ON spa.apparat_id = oda.apparat_id").
			Where("spa.system_part_id = ?", *systemPartID)
		query = query.Where("id IN (?)", sub)
	}

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(description) LIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ObjectData
	if err := r.withObjectDataPreloads(query).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func NewObjectDataRepository(db *gorm.DB) domainFacility.ObjectDataStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(description) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.ObjectData](db, searchCallback)
	return &objectDataRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *objectDataRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	if err := r.withObjectDataPreloads(r.db.Where("id IN ?", ids)).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *objectDataRepo) GetByID(id uuid.UUID) (*domainFacility.ObjectData, error) {
	var item domainFacility.ObjectData
	if err := r.withObjectDataPreloads(r.db.Where("id = ?", id)).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *objectDataRepo) Create(entity *domainFacility.ObjectData) error {
	// Mirror BaseRepository.Create behavior, but ensure Apparats association is saved.
	now := time.Now().UTC()
	if err := entity.GetBase().InitForCreate(now); err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		// Replace works for both empty and non-empty slices
		if entity.Apparats != nil {
			if err := tx.Model(entity).Association("Apparats").Replace(entity.Apparats); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *objectDataRepo) Update(entity *domainFacility.ObjectData) error {
	// Mirror BaseRepository.Update behavior (Save) and sync Apparats association.
	entity.GetBase().TouchForUpdate(time.Now().UTC())
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(entity).Error; err != nil {
			return err
		}
		if err := tx.Model(entity).Association("Apparats").Replace(entity.Apparats); err != nil {
			return err
		}
		return nil
	})
}

func (r *objectDataRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ObjectData{}).
		Where("project_id IS NULL")

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(description) LIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ObjectData
	if err := r.withObjectDataPreloads(query).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *objectDataRepo) GetBacnetObjectIDs(objectDataID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("object_data_bacnet_objects").
		Select("bacnet_object_id").
		Where("object_data_id = ?", objectDataID).
		Scan(&ids).Error
	return ids, err
}

func (r *objectDataRepo) ExistsByDescription(projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	desc := strings.ToLower(strings.TrimSpace(description))
	if desc == "" {
		return false, nil
	}

	query := r.db.Model(&domainFacility.ObjectData{})
	if projectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *projectID)
	}

	query = query.Where("LOWER(description) = ?", desc)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *objectDataRepo) GetTemplates() ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataPreloads(r.db.Where("is_active = ? AND project_id IS NULL", true)).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetForProject(projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataPreloads(r.db.Where("is_active = ? AND project_id = ?", true, projectID)).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetPaginatedListForProject(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(&projectID, nil, nil, params)
}

func (r *objectDataRepo) GetPaginatedListByApparatID(apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(nil, &apparatID, nil, params)
}

func (r *objectDataRepo) GetPaginatedListBySystemPartID(systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(nil, nil, &systemPartID, params)
}

func (r *objectDataRepo) GetPaginatedListByApparatAndSystemPartID(apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(nil, &apparatID, &systemPartID, params)
}

func (r *objectDataRepo) GetPaginatedListForProjectByApparatID(projectID, apparatID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(&projectID, &apparatID, nil, params)
}

func (r *objectDataRepo) GetPaginatedListForProjectBySystemPartID(projectID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(&projectID, nil, &systemPartID, params)
}

func (r *objectDataRepo) GetPaginatedListForProjectByApparatAndSystemPartID(projectID, apparatID, systemPartID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.getPaginatedListFiltered(&projectID, &apparatID, &systemPartID, params)
}
