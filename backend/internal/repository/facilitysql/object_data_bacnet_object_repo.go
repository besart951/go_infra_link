package facilitysql

import (
	"context"

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

func (r *objectDataBacnetObjectRepo) Add(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("object_data_bacnet_objects").Create(map[string]any{
		"object_data_id":   objectDataID,
		"bacnet_object_id": bacnetObjectID,
	}).Error
}

func (r *objectDataBacnetObjectRepo) Delete(ctx context.Context, objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("object_data_bacnet_objects").
		Where("object_data_id = ? AND bacnet_object_id = ?", objectDataID, bacnetObjectID).
		Delete(nil).Error
}

func (r *objectDataBacnetObjectRepo) DeleteByObjectDataID(ctx context.Context, objectDataID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("object_data_bacnet_objects").
		Where("object_data_id = ?", objectDataID).
		Delete(nil).Error
}
