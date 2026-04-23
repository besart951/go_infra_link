package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type fieldDeviceRecord struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 int
	TextIndividuell           *string
	SPSControllerSystemTypeID uuid.UUID                         `gorm:"type:uuid;not null;index"`
	SPSControllerSystemType   domainFacility.SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SystemPartID              uuid.UUID                         `gorm:"type:uuid;not null;index"`
	SystemPart                domainFacility.SystemPart         `gorm:"foreignKey:SystemPartID"`
	SpecificationID           *uuid.UUID                        `gorm:"type:uuid;index"`
	Specification             *domainFacility.Specification     `gorm:"foreignKey:SpecificationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ApparatID                 uuid.UUID                         `gorm:"type:uuid;not null;index"`
	Apparat                   domainFacility.Apparat            `gorm:"foreignKey:ApparatID"`
	BacnetObjects             []domainFacility.BacnetObject     `gorm:"foreignKey:FieldDeviceID"`
}

func (fieldDeviceRecord) TableName() string {
	return "field_devices"
}

func AutoMigrateFieldDeviceModels() []any {
	return []any{&fieldDeviceRecord{}}
}

func toFieldDeviceRecord(entity *domainFacility.FieldDevice) *fieldDeviceRecord {
	if entity == nil {
		return nil
	}

	return &fieldDeviceRecord{
		Base:                      entity.Base,
		BMK:                       entity.BMK,
		Description:               entity.Description,
		ApparatNr:                 entity.ApparatNr,
		TextIndividuell:           entity.TextIndividuell,
		SPSControllerSystemTypeID: entity.SPSControllerSystemTypeID,
		SystemPartID:              entity.SystemPartID,
		SpecificationID:           entity.SpecificationID,
		ApparatID:                 entity.ApparatID,
	}
}

func toFieldDeviceDomain(record *fieldDeviceRecord) *domainFacility.FieldDevice {
	if record == nil {
		return nil
	}

	return &domainFacility.FieldDevice{
		Base:                      record.Base,
		BMK:                       record.BMK,
		Description:               record.Description,
		ApparatNr:                 record.ApparatNr,
		TextIndividuell:           record.TextIndividuell,
		SPSControllerSystemTypeID: record.SPSControllerSystemTypeID,
		SPSControllerSystemType:   record.SPSControllerSystemType,
		SystemPartID:              record.SystemPartID,
		SystemPart:                record.SystemPart,
		SpecificationID:           record.SpecificationID,
		Specification:             record.Specification,
		ApparatID:                 record.ApparatID,
		Apparat:                   record.Apparat,
		BacnetObjects:             record.BacnetObjects,
	}
}

func toFieldDeviceDomains(records []*fieldDeviceRecord) []*domainFacility.FieldDevice {
	items := make([]*domainFacility.FieldDevice, len(records))
	for i, record := range records {
		items[i] = toFieldDeviceDomain(record)
	}
	return items
}

func fieldDeviceDomainValues(records []fieldDeviceRecord) []domainFacility.FieldDevice {
	items := make([]domainFacility.FieldDevice, len(records))
	for i := range records {
		items[i] = *toFieldDeviceDomain(&records[i])
	}
	return items
}