package project

import "github.com/google/uuid"

const (
	TaskCopyProjectControlCabinet          = "project.controlcabinet.copy.v1"
	TaskCopyProjectSPSController           = "project.spscontroller.copy.v1"
	TaskCopyProjectSPSControllerSystemType = "project.spscontrollersystemtype.copy.v1"
)

// FacilityCopyJobPayload is the durable command for project-scoped hierarchy
// copies. It contains identifiers only; the worker reloads current source data.
type FacilityCopyJobPayload struct {
	ProjectID uuid.UUID `json:"project_id"`
	SourceID  uuid.UUID `json:"source_id"`
}
