package facility

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

const (
	FacilityJobTaskMultiCreateFieldDevices = "fielddevice.multi_create.v1"
	FacilityJobTaskBulkUpdateFieldDevices  = "fielddevice.bulk_update.v1"
	FacilityJobTaskBulkDeleteFieldDevices  = "fielddevice.bulk_delete.v1"
)

type FieldDeviceMultiCreateTaskPayload struct {
	Items []domainFacility.FieldDeviceCreateItem `json:"items"`
}

type FieldDeviceBulkUpdateTaskPayload struct {
	Updates []domainFacility.BulkFieldDeviceUpdate `json:"updates"`
}

type FieldDeviceBulkDeleteTaskPayload struct {
	IDs      []uuid.UUID                               `json:"ids,omitempty"`
	Commands []domainFacility.FieldDeviceDeleteCommand `json:"commands,omitempty"`
}

type FieldDeviceBulkJobResult struct {
	TotalCount   int                         `json:"total_count"`
	SuccessCount int                         `json:"success_count"`
	FailureCount int                         `json:"failure_count"`
	Failures     []FieldDeviceBulkJobFailure `json:"failures,omitempty"`
}

type FieldDeviceBulkJobFailure struct {
	Ordinal  int       `json:"ordinal"`
	SourceID uuid.UUID `json:"source_id"`
	Error    string    `json:"error"`
}
