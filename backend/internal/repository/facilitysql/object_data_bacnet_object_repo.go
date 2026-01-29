package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataBacnetObjectRepo struct {
	db *gorm.DB
}

func NewObjectDataBacnetObjectRepository(db *gorm.DB) facility.ObjectDataBacnetObjectStore {
	return &objectDataBacnetObjectRepo{db: db}
}

func (r *objectDataBacnetObjectRepo) Add(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	return r.db.Table("object_data_bacnet_objects").Create(map[string]any{
		"object_data_id":   objectDataID,
		"bacnet_object_id": bacnetObjectID,
	}).Error
}

func (r *objectDataBacnetObjectRepo) Delete(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	return r.db.Table("object_data_bacnet_objects").
		Where("object_data_id = ? AND bacnet_object_id = ?", objectDataID, bacnetObjectID).
		Delete(nil).Error
}

func (r *objectDataBacnetObjectRepo) DeleteByObjectDataID(objectDataID uuid.UUID) error {
	return r.db.Table("object_data_bacnet_objects").
		Where("object_data_id = ?", objectDataID).
		Delete(nil).Error
}
