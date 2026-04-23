package projectsql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type projectControlCabinetRecord struct {
	domain.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
	ControlCabinetID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
}

func (projectControlCabinetRecord) TableName() string {
	return "project_control_cabinets"
}

type projectSPSControllerRecord struct {
	domain.Base
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
	SPSControllerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
}

func (projectSPSControllerRecord) TableName() string {
	return "project_sps_controllers"
}

type projectFieldDeviceRecord struct {
	domain.Base
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
	FieldDeviceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
}

func (projectFieldDeviceRecord) TableName() string {
	return "project_field_devices"
}

func AutoMigrateModels() []any {
	return []any{
		&projectControlCabinetRecord{},
		&projectSPSControllerRecord{},
		&projectFieldDeviceRecord{},
	}
}

func toProjectControlCabinetRecord(entity *domainProject.ProjectControlCabinet) *projectControlCabinetRecord {
	if entity == nil {
		return nil
	}
	return &projectControlCabinetRecord{
		Base:             entity.Base,
		ProjectID:        entity.ProjectID,
		ControlCabinetID: entity.ControlCabinetID,
	}
}

func toProjectControlCabinetDomain(record *projectControlCabinetRecord) *domainProject.ProjectControlCabinet {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectControlCabinet{
		Base:             record.Base,
		ProjectID:        record.ProjectID,
		ControlCabinetID: record.ControlCabinetID,
	}
}

func toProjectControlCabinetDomains(records []*projectControlCabinetRecord) []*domainProject.ProjectControlCabinet {
	items := make([]*domainProject.ProjectControlCabinet, len(records))
	for i, record := range records {
		items[i] = toProjectControlCabinetDomain(record)
	}
	return items
}

func projectControlCabinetDomainValues(records []projectControlCabinetRecord) []domainProject.ProjectControlCabinet {
	items := make([]domainProject.ProjectControlCabinet, len(records))
	for i := range records {
		items[i] = *toProjectControlCabinetDomain(&records[i])
	}
	return items
}

func toProjectSPSControllerRecord(entity *domainProject.ProjectSPSController) *projectSPSControllerRecord {
	if entity == nil {
		return nil
	}
	return &projectSPSControllerRecord{
		Base:            entity.Base,
		ProjectID:       entity.ProjectID,
		SPSControllerID: entity.SPSControllerID,
	}
}

func toProjectSPSControllerDomain(record *projectSPSControllerRecord) *domainProject.ProjectSPSController {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectSPSController{
		Base:            record.Base,
		ProjectID:       record.ProjectID,
		SPSControllerID: record.SPSControllerID,
	}
}

func toProjectSPSControllerDomains(records []*projectSPSControllerRecord) []*domainProject.ProjectSPSController {
	items := make([]*domainProject.ProjectSPSController, len(records))
	for i, record := range records {
		items[i] = toProjectSPSControllerDomain(record)
	}
	return items
}

func projectSPSControllerDomainValues(records []projectSPSControllerRecord) []domainProject.ProjectSPSController {
	items := make([]domainProject.ProjectSPSController, len(records))
	for i := range records {
		items[i] = *toProjectSPSControllerDomain(&records[i])
	}
	return items
}

func toProjectFieldDeviceRecord(entity *domainProject.ProjectFieldDevice) *projectFieldDeviceRecord {
	if entity == nil {
		return nil
	}
	return &projectFieldDeviceRecord{
		Base:          entity.Base,
		ProjectID:     entity.ProjectID,
		FieldDeviceID: entity.FieldDeviceID,
	}
}

func toProjectFieldDeviceDomain(record *projectFieldDeviceRecord) *domainProject.ProjectFieldDevice {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectFieldDevice{
		Base:          record.Base,
		ProjectID:     record.ProjectID,
		FieldDeviceID: record.FieldDeviceID,
	}
}

func toProjectFieldDeviceDomains(records []*projectFieldDeviceRecord) []*domainProject.ProjectFieldDevice {
	items := make([]*domainProject.ProjectFieldDevice, len(records))
	for i, record := range records {
		items[i] = toProjectFieldDeviceDomain(record)
	}
	return items
}

func projectFieldDeviceDomainValues(records []projectFieldDeviceRecord) []domainProject.ProjectFieldDevice {
	items := make([]domainProject.ProjectFieldDevice, len(records))
	for i := range records {
		items[i] = *toProjectFieldDeviceDomain(&records[i])
	}
	return items
}
