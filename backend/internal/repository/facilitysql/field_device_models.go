package facilitysql

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type FieldDeviceRecord struct {
	domain.Base
	BMK                       *string
	Description               *string
	ApparatNr                 int
	TextIndividuell           *string
	SPSControllerSystemTypeID uuid.UUID                              `gorm:"type:uuid;not null;index"`
	SPSControllerSystemType   domainFacility.SPSControllerSystemType `gorm:"foreignKey:SPSControllerSystemTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SystemPartID              uuid.UUID                              `gorm:"type:uuid;not null;index"`
	SystemPart                domainFacility.SystemPart              `gorm:"foreignKey:SystemPartID"`
	SpecificationID           *uuid.UUID                             `gorm:"type:uuid;index"`
	Specification             *domainFacility.Specification          `gorm:"foreignKey:SpecificationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ApparatID                 uuid.UUID                              `gorm:"type:uuid;not null;index"`
	Apparat                   domainFacility.Apparat                 `gorm:"foreignKey:ApparatID"`
	BacnetObjects             []domainFacility.BacnetObject          `gorm:"foreignKey:FieldDeviceID"`
}

func (FieldDeviceRecord) TableName() string {
	return "field_devices"
}

func toFieldDeviceRecord(entity *domainFacility.FieldDevice) *FieldDeviceRecord {
	if entity == nil {
		return nil
	}

	return &FieldDeviceRecord{
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

func toFieldDeviceDomain(record *FieldDeviceRecord) *domainFacility.FieldDevice {
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

func toFieldDeviceDomains(records []*FieldDeviceRecord) []*domainFacility.FieldDevice {
	items := make([]*domainFacility.FieldDevice, len(records))
	for i, record := range records {
		items[i] = toFieldDeviceDomain(record)
	}
	return items
}

type fieldDeviceListRow struct {
	ID                        uuid.UUID  `gorm:"column:id"`
	CreatedAt                 time.Time  `gorm:"column:created_at"`
	UpdatedAt                 time.Time  `gorm:"column:updated_at"`
	BMK                       *string    `gorm:"column:bmk"`
	Description               *string    `gorm:"column:description"`
	ApparatNr                 int        `gorm:"column:apparat_nr"`
	TextIndividuell           *string    `gorm:"column:text_individuell"`
	SPSControllerSystemTypeID uuid.UUID  `gorm:"column:sps_controller_system_type_id"`
	SystemPartID              uuid.UUID  `gorm:"column:system_part_id"`
	SpecificationID           *uuid.UUID `gorm:"column:specification_id"`
	ApparatID                 uuid.UUID  `gorm:"column:apparat_id"`

	SPSSystemTypeID         *uuid.UUID `gorm:"column:sps_system_type_id"`
	SPSSystemTypeCreatedAt  *time.Time `gorm:"column:sps_system_type_created_at"`
	SPSSystemTypeUpdatedAt  *time.Time `gorm:"column:sps_system_type_updated_at"`
	SPSSystemTypeNumber     *int       `gorm:"column:sps_system_type_number"`
	SPSSystemTypeDocument   *string    `gorm:"column:sps_system_type_document_name"`
	SPSControllerID         *uuid.UUID `gorm:"column:sps_controller_id"`
	SPSControllerDeviceName *string    `gorm:"column:sps_controller_device_name"`
	SystemTypeID            *uuid.UUID `gorm:"column:system_type_id"`
	SystemTypeName          *string    `gorm:"column:system_type_name"`
	ApparatListID           *uuid.UUID `gorm:"column:apparat_list_id"`
	ApparatCreatedAt        *time.Time `gorm:"column:apparat_created_at"`
	ApparatUpdatedAt        *time.Time `gorm:"column:apparat_updated_at"`
	ApparatShortName        *string    `gorm:"column:apparat_short_name"`
	ApparatName             *string    `gorm:"column:apparat_name"`
	ApparatDescription      *string    `gorm:"column:apparat_description"`
	SystemPartListID        *uuid.UUID `gorm:"column:system_part_list_id"`
	SystemPartCreatedAt     *time.Time `gorm:"column:system_part_created_at"`
	SystemPartUpdatedAt     *time.Time `gorm:"column:system_part_updated_at"`
	SystemPartShortName     *string    `gorm:"column:system_part_short_name"`
	SystemPartName          *string    `gorm:"column:system_part_name"`
	SystemPartDescription   *string    `gorm:"column:system_part_description"`
}

func (row fieldDeviceListRow) toDomain() domainFacility.FieldDevice {
	item := domainFacility.FieldDevice{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		BMK:                       row.BMK,
		Description:               row.Description,
		ApparatNr:                 row.ApparatNr,
		TextIndividuell:           row.TextIndividuell,
		SPSControllerSystemTypeID: row.SPSControllerSystemTypeID,
		SystemPartID:              row.SystemPartID,
		SpecificationID:           row.SpecificationID,
		ApparatID:                 row.ApparatID,
	}

	if row.SPSSystemTypeID != nil {
		spsSystemType := domainFacility.SPSControllerSystemType{
			Base: domain.Base{
				ID: *row.SPSSystemTypeID,
			},
			Number:       row.SPSSystemTypeNumber,
			DocumentName: row.SPSSystemTypeDocument,
		}
		if row.SPSSystemTypeCreatedAt != nil {
			spsSystemType.CreatedAt = *row.SPSSystemTypeCreatedAt
		}
		if row.SPSSystemTypeUpdatedAt != nil {
			spsSystemType.UpdatedAt = *row.SPSSystemTypeUpdatedAt
		}
		if row.SPSControllerID != nil {
			spsSystemType.SPSControllerID = *row.SPSControllerID
			spsSystemType.SPSController = domainFacility.SPSController{
				Base: domain.Base{ID: *row.SPSControllerID},
			}
			if row.SPSControllerDeviceName != nil {
				spsSystemType.SPSController.DeviceName = *row.SPSControllerDeviceName
			}
		}
		if row.SystemTypeID != nil {
			spsSystemType.SystemTypeID = *row.SystemTypeID
			spsSystemType.SystemType = domainFacility.SystemType{
				Base: domain.Base{ID: *row.SystemTypeID},
			}
			if row.SystemTypeName != nil {
				spsSystemType.SystemType.Name = *row.SystemTypeName
			}
		}
		item.SPSControllerSystemType = spsSystemType
	}

	if row.ApparatListID != nil {
		apparat := domainFacility.Apparat{
			Base: domain.Base{ID: *row.ApparatListID},
		}
		if row.ApparatCreatedAt != nil {
			apparat.CreatedAt = *row.ApparatCreatedAt
		}
		if row.ApparatUpdatedAt != nil {
			apparat.UpdatedAt = *row.ApparatUpdatedAt
		}
		if row.ApparatShortName != nil {
			apparat.ShortName = *row.ApparatShortName
		}
		if row.ApparatName != nil {
			apparat.Name = *row.ApparatName
		}
		apparat.Description = row.ApparatDescription
		item.Apparat = apparat
	}

	if row.SystemPartListID != nil {
		systemPart := domainFacility.SystemPart{
			Base: domain.Base{ID: *row.SystemPartListID},
		}
		if row.SystemPartCreatedAt != nil {
			systemPart.CreatedAt = *row.SystemPartCreatedAt
		}
		if row.SystemPartUpdatedAt != nil {
			systemPart.UpdatedAt = *row.SystemPartUpdatedAt
		}
		if row.SystemPartShortName != nil {
			systemPart.ShortName = *row.SystemPartShortName
		}
		if row.SystemPartName != nil {
			systemPart.Name = *row.SystemPartName
		}
		systemPart.Description = row.SystemPartDescription
		item.SystemPart = systemPart
	}

	return item
}

func fieldDeviceListRowDomainValues(rows []fieldDeviceListRow) []domainFacility.FieldDevice {
	items := make([]domainFacility.FieldDevice, len(rows))
	for i := range rows {
		items[i] = rows[i].toDomain()
	}
	return items
}
