package facilitysql

import (
	"context"
	"time"

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

func (r *bacnetObjectTemplateRepo) ListByObjectDataID(ctx context.Context, objectDataID uuid.UUID) ([]domainObjectData.BacnetObjectTemplate, error) {
	var records []BacnetObjectTemplateRecord
	err := r.db.WithContext(ctx).Where("object_data_id = ?", objectDataID).
		Order("software_type ASC, software_number ASC, id ASC").Find(&records).Error
	items := make([]domainObjectData.BacnetObjectTemplate, len(records))
	for i := range records {
		items[i] = templateRecordToDomain(records[i])
	}
	return items, err
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
		return tx.CreateInBatches(records, 500).Error
	})
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
