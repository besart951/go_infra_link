package facilitysql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectTemplateRepo struct {
	db *gorm.DB
}

func NewBacnetObjectTemplateRepository(db *gorm.DB) domainObjectData.BacnetObjectTemplateStore {
	return &bacnetObjectTemplateRepo{db: db}
}

func (r *bacnetObjectTemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainObjectData.BacnetObjectTemplate, error) {
	items, err := r.GetByIDs(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, domain.ErrNotFound
	}
	return &items[0], nil
}

func (r *bacnetObjectTemplateRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domainObjectData.BacnetObjectTemplate, error) {
	if len(ids) == 0 {
		return []domainObjectData.BacnetObjectTemplate{}, nil
	}
	var records []BacnetObjectTemplateRecord
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error; err != nil {
		return nil, err
	}
	return r.withAlarmValues(ctx, records)
}

func (r *bacnetObjectTemplateRepo) ListByObjectDataID(ctx context.Context, objectDataID uuid.UUID) ([]domainObjectData.BacnetObjectTemplate, error) {
	var records []BacnetObjectTemplateRecord
	err := r.db.WithContext(ctx).Where("object_data_id = ?", objectDataID).
		Order("software_type ASC, software_number ASC, id ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return r.withAlarmValues(ctx, records)
}

func (r *bacnetObjectTemplateRepo) Create(ctx context.Context, template *domainObjectData.BacnetObjectTemplate) error {
	if template == nil {
		return domain.ErrInvalidArgument
	}
	prepared, err := prepareTemplateRecords(template.ObjectDataID, []domainObjectData.BacnetObjectTemplate{*template})
	if err != nil {
		return err
	}
	alarmValues := template.AlarmValues
	*template = templateRecordToDomain(prepared[0])
	template.AlarmValues = alarmValues
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&prepared[0]).Error; err != nil {
			return err
		}
		return replaceTemplateAlarmValues(tx, template.ID, template.AlarmValues)
	})
}

func (r *bacnetObjectTemplateRepo) Update(ctx context.Context, template *domainObjectData.BacnetObjectTemplate) error {
	if template == nil || template.Version == 0 {
		return domain.ErrInvalidArgument
	}
	expected := template.Version
	template.Version++
	template.UpdatedAt = time.Now().UTC()
	record := templateRecordFromDomain(*template)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&BacnetObjectTemplateRecord{}).Where("id=? AND version=?", template.ID, expected).
			Select("*").Omit("id", "created_at").Updates(&record)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			template.Version = expected
			return domain.ErrConflict
		}
		return replaceTemplateAlarmValues(tx, template.ID, template.AlarmValues)
	})
}

func (r *bacnetObjectTemplateRepo) DeleteAtVersion(ctx context.Context, id uuid.UUID, version uint64) error {
	if id == uuid.Nil || version == 0 {
		return domain.ErrInvalidArgument
	}
	result := r.db.WithContext(ctx).Where("id=? AND version=?", id, version).Delete(&BacnetObjectTemplateRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *bacnetObjectTemplateRepo) Replace(ctx context.Context, objectDataID uuid.UUID, templates []domainObjectData.BacnetObjectTemplate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("object_data_id = ?", objectDataID).Delete(&BacnetObjectTemplateRecord{}).Error; err != nil {
			return err
		}
		records, err := prepareTemplateRecords(objectDataID, templates)
		if err != nil || len(records) == 0 {
			return err
		}
		if err := tx.CreateInBatches(records, 500).Error; err != nil {
			return err
		}
		for i := range templates {
			if err := replaceTemplateAlarmValues(tx, records[i].ID, templates[i].AlarmValues); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *bacnetObjectTemplateRepo) withAlarmValues(ctx context.Context, records []BacnetObjectTemplateRecord) ([]domainObjectData.BacnetObjectTemplate, error) {
	items := make([]domainObjectData.BacnetObjectTemplate, len(records))
	ids := make([]uuid.UUID, len(records))
	for i := range records {
		items[i], ids[i] = templateRecordToDomain(records[i]), records[i].ID
	}
	var values []BacnetObjectTemplateAlarmValueRecord
	if len(ids) > 0 {
		if err := r.db.WithContext(ctx).Where("template_id IN ?", ids).Order("template_id,id").Find(&values).Error; err != nil {
			return nil, err
		}
	}
	byID := make(map[uuid.UUID]int, len(items))
	for i := range items {
		byID[items[i].ID] = i
	}
	for _, value := range values {
		index := byID[value.TemplateID]
		items[index].AlarmValues = append(items[index].AlarmValues, templateAlarmRecordToDomain(value))
	}
	return items, nil
}

func replaceTemplateAlarmValues(tx *gorm.DB, templateID uuid.UUID, values []domainObjectData.BacnetObjectTemplateAlarmValue) error {
	if err := tx.Where("template_id=?", templateID).Delete(&BacnetObjectTemplateAlarmValueRecord{}).Error; err != nil {
		return err
	}
	now := time.Now().UTC()
	records := make([]BacnetObjectTemplateAlarmValueRecord, len(values))
	for i := range values {
		value := values[i]
		value.TemplateID = templateID
		if value.ID == uuid.Nil {
			value.ID = uuid.New()
		}
		if value.CreatedAt.IsZero() {
			value.CreatedAt, value.Version = now, 1
		}
		value.UpdatedAt = now
		records[i] = templateAlarmRecordFromDomain(value)
	}
	if len(records) == 0 {
		return nil
	}
	return tx.CreateInBatches(records, 500).Error
}

func prepareTemplateRecords(objectDataID uuid.UUID, templates []domainObjectData.BacnetObjectTemplate) ([]BacnetObjectTemplateRecord, error) {
	now := time.Now().UTC()
	records := make([]BacnetObjectTemplateRecord, len(templates))
	for i := range templates {
		template := templates[i]
		template.ObjectDataID = objectDataID
		if template.ID == uuid.Nil {
			template.ID = uuid.New()
		}
		if template.CreatedAt.IsZero() {
			template.CreatedAt, template.Version = now, 1
		}
		template.UpdatedAt = now
		records[i] = templateRecordFromDomain(template)
	}
	return records, nil
}
