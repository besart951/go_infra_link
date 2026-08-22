package project

import "github.com/google/uuid"

const (
	TaskCopyProjectControlCabinet          = "project.controlcabinet.copy.v1"
	TaskCopyProjectSPSController           = "project.spscontroller.copy.v1"
	TaskCopyProjectSPSControllerSystemType = "project.spscontrollersystemtype.copy.v1"
)

// ProjectFacilityCopyCommand is the durable command for project-scoped hierarchy
// copies. It contains identifiers only; the worker reloads current source data.
type ProjectFacilityCopyCommand struct {
	ProjectID uuid.UUID `json:"project_id"`
	SourceID  uuid.UUID `json:"source_id"`
}
