package facilitysql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataRepo struct {
	*gormbase.BaseRepository[*domainFacility.ObjectData]
	db *gorm.DB
}

type fieldDeviceOptionsRevisionEntity struct {
	ID        uuid.UUID `gorm:"column:id"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type fieldDeviceOptionsObjectDataApparatRevision struct {
	ObjectDataID uuid.UUID `gorm:"column:object_data_id"`
	ApparatID    uuid.UUID `gorm:"column:apparat_id"`
}

type fieldDeviceOptionsSystemPartApparatRevision struct {
	SystemPartID uuid.UUID `gorm:"column:system_part_id"`
	ApparatID    uuid.UUID `gorm:"column:apparat_id"`
}

func orderFieldDeviceOptionObjectDatas(query *gorm.DB) *gorm.DB {
	return query.
		Order("LOWER(description) ASC").
		Order("LOWER(obj_version) ASC").
		Order("id ASC")
}

func orderFieldDeviceOptionApparats(query *gorm.DB) *gorm.DB {
	return query.
		Order("LOWER(short_name) ASC").
		Order("LOWER(name) ASC").
		Order("id ASC")
}

func (r *objectDataRepo) withObjectDataPreloads(query *gorm.DB) *gorm.DB {
	return query.
		Preload("BacnetObjects").
		Preload("Apparats", func(db *gorm.DB) *gorm.DB {
			return orderFieldDeviceOptionApparats(db)
		})
}

func (r *objectDataRepo) withObjectDataLitePreloads(query *gorm.DB) *gorm.DB {
	return query.Preload("Apparats", func(db *gorm.DB) *gorm.DB {
		return orderFieldDeviceOptionApparats(db)
	})
}

func (r *objectDataRepo) GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.ObjectDataFilterParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainFacility.ObjectData{})
	if filters.ProjectID == nil {
		query = query.Where("project_id IS NULL")
	} else {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}

	if filters.ApparatID != nil {
		sub := r.db.WithContext(ctx).Table("object_data_apparats").
			Select("object_data_id").
			Where("apparat_id = ?", *filters.ApparatID)
		query = query.Where("id IN (?)", sub)
	}

	if filters.SystemPartID != nil {
		sub := r.db.WithContext(ctx).Table("object_data_apparats AS oda").
			Select("DISTINCT oda.object_data_id").
			Joins("JOIN system_part_apparats spa ON spa.apparat_id = oda.apparat_id").
			Where("spa.system_part_id = ?", *filters.SystemPartID)
		query = query.Where("id IN (?)", sub)
	}

	if strings.TrimSpace(params.Search) != "" {
		query = applyObjectDataSearch(query, params.Search)
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
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.ObjectData](searchspec.ObjectData.SearchColumns("")...),
	)
	return &objectDataRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *objectDataRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	if err := r.withObjectDataPreloads(r.db.WithContext(ctx).Where("id IN ?", ids)).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *objectDataRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.ObjectData, error) {
	var item domainFacility.ObjectData
	if err := r.withObjectDataPreloads(r.db.WithContext(ctx).Where("id = ?", id)).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *objectDataRepo) Create(ctx context.Context, entity *domainFacility.ObjectData) error {
	// Mirror BaseRepository.Create behavior, but ensure Apparats association is saved.
	now := time.Now().UTC()
	if err := entity.GetBase().InitForCreate(now); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *objectDataRepo) Update(ctx context.Context, entity *domainFacility.ObjectData) error {
	// Mirror BaseRepository.Update behavior (Save) and sync Apparats association.
	entity.GetBase().TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(entity).Error; err != nil {
			return err
		}
		if err := tx.Model(entity).Association("Apparats").Replace(entity.Apparats); err != nil {
			return err
		}
		return nil
	})
}

func (r *objectDataRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return r.GetPaginatedListWithFilters(ctx, params, domainFacility.ObjectDataFilterParams{})
}

func applyObjectDataSearch(query *gorm.DB, search string) *gorm.DB {
	return gormbase.ApplyTrigramSearch(query, search, searchspec.ObjectData.SearchColumns("")...)
}

func (r *objectDataRepo) GetBacnetObjectIDs(ctx context.Context, objectDataID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("object_data_bacnet_objects").
		Select("bacnet_object_id").
		Where("object_data_id = ?", objectDataID).
		Scan(&ids).Error
	return ids, err
}

func (r *objectDataRepo) ExistsByDescription(ctx context.Context, projectID *uuid.UUID, description string, excludeID *uuid.UUID) (bool, error) {
	desc := strings.ToLower(strings.TrimSpace(description))
	if desc == "" {
		return false, nil
	}

	query := r.db.WithContext(ctx).Model(&domainFacility.ObjectData{})
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

func (r *objectDataRepo) GetTemplates(ctx context.Context) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataPreloads(
		orderFieldDeviceOptionObjectDatas(
			r.db.WithContext(ctx).Where("is_active = ? AND project_id IS NULL", true),
		),
	).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetTemplatesLite(ctx context.Context) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataLitePreloads(
		orderFieldDeviceOptionObjectDatas(
			r.db.WithContext(ctx).Where("is_active = ? AND project_id IS NULL", true),
		),
	).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetForProject(ctx context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataPreloads(
		orderFieldDeviceOptionObjectDatas(
			r.db.WithContext(ctx).Where("is_active = ? AND project_id = ?", true, projectID),
		),
	).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetForProjectLite(ctx context.Context, projectID uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.withObjectDataLitePreloads(
		orderFieldDeviceOptionObjectDatas(
			r.db.WithContext(ctx).Where("is_active = ? AND project_id = ?", true, projectID),
		),
	).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetFieldDeviceOptionsRevision(ctx context.Context, projectID *uuid.UUID) (string, error) {
	var objectDatas []fieldDeviceOptionsRevisionEntity
	objectDataQuery := r.db.WithContext(ctx).
		Model(&domainFacility.ObjectData{}).
		Select("id, updated_at").
		Where("is_active = ?", true)
	if projectID == nil {
		objectDataQuery = objectDataQuery.Where("project_id IS NULL")
	} else {
		objectDataQuery = objectDataQuery.Where("project_id = ?", *projectID)
	}
	if err := objectDataQuery.Find(&objectDatas).Error; err != nil {
		return "", err
	}

	objectDataIDs := revisionEntityIDs(objectDatas)
	h := sha256.New()
	writeRevisionEntities(h, "object_data", objectDatas)

	objectDataApparats := []fieldDeviceOptionsObjectDataApparatRevision{}
	if len(objectDataIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Table("object_data_apparats").
			Select("object_data_id, apparat_id").
			Where("object_data_id IN ?", objectDataIDs).
			Scan(&objectDataApparats).Error; err != nil {
			return "", err
		}
	}
	writeObjectDataApparatRevision(h, objectDataApparats)

	apparatIDs := objectDataApparatIDs(objectDataApparats)
	var apparats []fieldDeviceOptionsRevisionEntity
	if len(apparatIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Model(&domainFacility.Apparat{}).
			Select("id, updated_at").
			Where("id IN ?", apparatIDs).
			Find(&apparats).Error; err != nil {
			return "", err
		}
	}
	writeRevisionEntities(h, "apparat", apparats)

	systemPartApparats := []fieldDeviceOptionsSystemPartApparatRevision{}
	if len(apparatIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Table("system_part_apparats").
			Select("system_part_id, apparat_id").
			Where("apparat_id IN ?", apparatIDs).
			Scan(&systemPartApparats).Error; err != nil {
			return "", err
		}
	}
	writeSystemPartApparatRevision(h, systemPartApparats)

	systemPartIDs := systemPartApparatIDs(systemPartApparats)
	var systemParts []fieldDeviceOptionsRevisionEntity
	if len(systemPartIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Model(&domainFacility.SystemPart{}).
			Select("id, updated_at").
			Where("id IN ?", systemPartIDs).
			Find(&systemParts).Error; err != nil {
			return "", err
		}
	}
	writeRevisionEntities(h, "system_part", systemParts)

	return hex.EncodeToString(h.Sum(nil)), nil
}

func revisionEntityIDs(rows []fieldDeviceOptionsRevisionEntity) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func objectDataApparatIDs(rows []fieldDeviceOptionsObjectDataApparatRevision) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.ApparatID]; ok {
			continue
		}
		seen[row.ApparatID] = struct{}{}
		ids = append(ids, row.ApparatID)
	}
	return ids
}

func systemPartApparatIDs(rows []fieldDeviceOptionsSystemPartApparatRevision) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.SystemPartID]; ok {
			continue
		}
		seen[row.SystemPartID] = struct{}{}
		ids = append(ids, row.SystemPartID)
	}
	return ids
}

func writeRevisionEntities(h hash.Hash, label string, rows []fieldDeviceOptionsRevisionEntity) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].ID.String() < rows[j].ID.String()
	})

	fmt.Fprintf(h, "%s:%d;", label, len(rows))
	for _, row := range rows {
		fmt.Fprintf(h, "%s:%s;", row.ID.String(), row.UpdatedAt.UTC().Format(time.RFC3339Nano))
	}
}

func writeObjectDataApparatRevision(h hash.Hash, rows []fieldDeviceOptionsObjectDataApparatRevision) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ObjectDataID == rows[j].ObjectDataID {
			return rows[i].ApparatID.String() < rows[j].ApparatID.String()
		}
		return rows[i].ObjectDataID.String() < rows[j].ObjectDataID.String()
	})

	fmt.Fprintf(h, "object_data_apparats:%d;", len(rows))
	for _, row := range rows {
		fmt.Fprintf(h, "%s:%s;", row.ObjectDataID.String(), row.ApparatID.String())
	}
}

func writeSystemPartApparatRevision(h hash.Hash, rows []fieldDeviceOptionsSystemPartApparatRevision) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ApparatID == rows[j].ApparatID {
			return rows[i].SystemPartID.String() < rows[j].SystemPartID.String()
		}
		return rows[i].ApparatID.String() < rows[j].ApparatID.String()
	})

	fmt.Fprintf(h, "system_part_apparats:%d;", len(rows))
	for _, row := range rows {
		fmt.Fprintf(h, "%s:%s;", row.ApparatID.String(), row.SystemPartID.String())
	}
}
