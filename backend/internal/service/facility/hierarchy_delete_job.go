package facility

import (
	"errors"

	"github.com/google/uuid"
)

const (
	FacilityJobTaskDeleteControlCabinet          = "controlcabinet.delete.v1"
	FacilityJobTaskDeleteSPSController           = "spscontroller.delete.v1"
	FacilityJobTaskDeleteSPSControllerSystemType = "spscontrollersystemtype.delete.v1"
)

type FacilityDeleteTaskPayload struct {
	SourceID uuid.UUID `json:"source_id"`
}

type FacilityAggregateState string

const (
	FacilityAggregateStateDeleting       FacilityAggregateState = "deleting"
	FacilityAggregateStateRestoreStaging FacilityAggregateState = "restore_staging"
)

var (
	ErrAggregateLocked   = errors.New("facility aggregate is locked")
	ErrAggregateNotFound = errors.New("facility aggregate not found")
)

// FacilityAggregateAdmission is persisted atomically with a mutation job.
// It keeps partially processed hierarchy graphs invisible and immutable.
type FacilityAggregateAdmission struct {
	ResourceID   uuid.UUID
	State        FacilityAggregateState
	AllowMissing bool
}
