package facilitysql

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type objectDataBacnetObjectRepo struct {
	db *gorm.DB
}

func NewObjectDataBacnetObjectRepository(db *gorm.DB) domainObjectData.ObjectDataBacnetObjectStore {
	return &objectDataBacnetObjectRepo{db: db}
}

func (r *objectDataBacnetObjectRepo) Add(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	db := r.db.WithContext(ctx)
	err := db.Table("object_data_bacnet_objects").Clauses(clause.OnConflict{DoNothing: true}).Create(map[string]any{
		"object_data_id":   objectDataID,
		"bacnet_object_id": bacnetObjectID,
	}).Error
	if err != nil || !db.Migrator().HasTable(&BacnetObjectTemplateRecord{}) {
		return err
	}
	return r.upsertTemplate(ctx, objectDataID, bacnetObjectID)
}

func (r *objectDataBacnetObjectRepo) Delete(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	db := r.db.WithContext(ctx)
	if db.Migrator().HasTable(&BacnetObjectTemplateRecord{}) {
		if err := db.Where("object_data_id = ? AND id = ?", objectDataID, bacnetObjectID).
			Delete(&BacnetObjectTemplateRecord{}).Error; err != nil {
			return err
		}
	}
	return db.Table("object_data_bacnet_objects").
		Where("object_data_id = ? AND bacnet_object_id = ?", objectDataID, bacnetObjectID).
		Delete(nil).Error
}

func (r *objectDataBacnetObjectRepo) DeleteByObjectDataID(ctx context.Context, objectDataID uuid.UUID) error {
	db := r.db.WithContext(ctx)
	if db.Migrator().HasTable(&BacnetObjectTemplateRecord{}) {
		if err := db.Where("object_data_id = ?", objectDataID).Delete(&BacnetObjectTemplateRecord{}).Error; err != nil {
			return err
		}
	}
	return db.Table("object_data_bacnet_objects").
		Where("object_data_id = ?", objectDataID).
		Delete(nil).Error
}

func (r *objectDataBacnetObjectRepo) upsertTemplate(ctx context.Context, objectDataID, bacnetObjectID uuid.UUID) error {
	var source domainFacility.BacnetObject
	if err := r.db.WithContext(ctx).First(&source, "id = ?", bacnetObjectID).Error; err != nil {
		return err
	}
	record := bacnetTemplateRecordFromLegacy(objectDataID, source)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&record).Error
}

func bacnetTemplateRecordFromLegacy(objectDataID uuid.UUID, source domainFacility.BacnetObject) BacnetObjectTemplateRecord {
	return BacnetObjectTemplateRecord{
		ID: source.ID, CreatedAt: source.CreatedAt, UpdatedAt: source.UpdatedAt, Version: source.Version,
		ObjectDataID: objectDataID, TextFix: source.TextFix, Description: source.Description,
		GMSVisible: source.GMSVisible, Optional: source.Optional, TextIndividual: source.TextIndividual,
		SoftwareType: source.SoftwareType, SoftwareNumber: source.SoftwareNumber,
		HardwareType: source.HardwareType, HardwareQuantity: source.HardwareQuantity,
		SoftwareReferenceID: source.SoftwareReferenceID, StateTextID: source.StateTextID,
		NotificationClassID: source.NotificationClassID, AlarmTypeID: source.AlarmTypeID,
	}
}
