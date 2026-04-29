package projectsql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ProjectControlCabinetRecord struct {
	domain.Base
	ProjectID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
	ControlCabinetID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_control_cabinet_unique"`
}

func (ProjectControlCabinetRecord) TableName() string {
	return "project_control_cabinets"
}

type ProjectSPSControllerRecord struct {
	domain.Base
	ProjectID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
	SPSControllerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_sps_controller_unique"`
}

func (ProjectSPSControllerRecord) TableName() string {
	return "project_sps_controllers"
}

type ProjectFieldDeviceRecord struct {
	domain.Base
	ProjectID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
	FieldDeviceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_field_device_unique"`
}

func (ProjectFieldDeviceRecord) TableName() string {
	return "project_field_devices"
}

func toProjectControlCabinetRecord(entity *domainProject.ProjectControlCabinet) *ProjectControlCabinetRecord {
	if entity == nil {
		return nil
	}
	return &ProjectControlCabinetRecord{
		Base:             entity.Base,
		ProjectID:        entity.ProjectID,
		ControlCabinetID: entity.ControlCabinetID,
	}
}

func toProjectControlCabinetDomain(record *ProjectControlCabinetRecord) *domainProject.ProjectControlCabinet {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectControlCabinet{
		Base:             record.Base,
		ProjectID:        record.ProjectID,
		ControlCabinetID: record.ControlCabinetID,
	}
}

func toProjectControlCabinetDomains(records []*ProjectControlCabinetRecord) []*domainProject.ProjectControlCabinet {
	items := make([]*domainProject.ProjectControlCabinet, len(records))
	for i, record := range records {
		items[i] = toProjectControlCabinetDomain(record)
	}
	return items
}

func projectControlCabinetDomainValues(records []ProjectControlCabinetRecord) []domainProject.ProjectControlCabinet {
	items := make([]domainProject.ProjectControlCabinet, len(records))
	for i := range records {
		items[i] = *toProjectControlCabinetDomain(&records[i])
	}
	return items
}

func toProjectSPSControllerRecord(entity *domainProject.ProjectSPSController) *ProjectSPSControllerRecord {
	if entity == nil {
		return nil
	}
	return &ProjectSPSControllerRecord{
		Base:            entity.Base,
		ProjectID:       entity.ProjectID,
		SPSControllerID: entity.SPSControllerID,
	}
}

func toProjectSPSControllerDomain(record *ProjectSPSControllerRecord) *domainProject.ProjectSPSController {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectSPSController{
		Base:            record.Base,
		ProjectID:       record.ProjectID,
		SPSControllerID: record.SPSControllerID,
	}
}

func toProjectSPSControllerDomains(records []*ProjectSPSControllerRecord) []*domainProject.ProjectSPSController {
	items := make([]*domainProject.ProjectSPSController, len(records))
	for i, record := range records {
		items[i] = toProjectSPSControllerDomain(record)
	}
	return items
}

func projectSPSControllerDomainValues(records []ProjectSPSControllerRecord) []domainProject.ProjectSPSController {
	items := make([]domainProject.ProjectSPSController, len(records))
	for i := range records {
		items[i] = *toProjectSPSControllerDomain(&records[i])
	}
	return items
}

func toProjectFieldDeviceRecord(entity *domainProject.ProjectFieldDevice) *ProjectFieldDeviceRecord {
	if entity == nil {
		return nil
	}
	return &ProjectFieldDeviceRecord{
		Base:          entity.Base,
		ProjectID:     entity.ProjectID,
		FieldDeviceID: entity.FieldDeviceID,
	}
}

func toProjectFieldDeviceDomain(record *ProjectFieldDeviceRecord) *domainProject.ProjectFieldDevice {
	if record == nil {
		return nil
	}
	return &domainProject.ProjectFieldDevice{
		Base:          record.Base,
		ProjectID:     record.ProjectID,
		FieldDeviceID: record.FieldDeviceID,
	}
}

func toProjectFieldDeviceDomains(records []*ProjectFieldDeviceRecord) []*domainProject.ProjectFieldDevice {
	items := make([]*domainProject.ProjectFieldDevice, len(records))
	for i, record := range records {
		items[i] = toProjectFieldDeviceDomain(record)
	}
	return items
}

func projectFieldDeviceDomainValues(records []ProjectFieldDeviceRecord) []domainProject.ProjectFieldDevice {
	items := make([]domainProject.ProjectFieldDevice, len(records))
	for i := range records {
		items[i] = *toProjectFieldDeviceDomain(&records[i])
	}
	return items
}
