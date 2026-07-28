package fielddevice

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// BulkUpdatePhase identifies one independently attempted part of a
// FieldDevice bulk update.
type BulkUpdatePhase string

const (
	BulkUpdatePhaseFieldDevice   BulkUpdatePhase = "fielddevice"
	BulkUpdatePhaseSpecification BulkUpdatePhase = "specification"
	BulkUpdatePhaseBacnetObjects BulkUpdatePhase = "bacnet_objects"
)

// BulkUpdatePhaseStatus reports whether the compatibility workflow completed a
// requested phase. Failed does not claim rollback: the legacy Specification
// and BACnet implementations can partially persist within a failed phase.
type BulkUpdatePhaseStatus string

const (
	BulkUpdatePhaseSucceeded    BulkUpdatePhaseStatus = "succeeded"
	BulkUpdatePhaseFailed       BulkUpdatePhaseStatus = "failed"
	BulkUpdatePhaseNotAttempted BulkUpdatePhaseStatus = "not_attempted"
)

type BulkUpdatePhaseResult struct {
	Phase  BulkUpdatePhase
	Status BulkUpdatePhaseStatus
}

// BulkUpdateItemExecution is index-aligned with the input. Index is retained
// because duplicate FieldDevice IDs are currently accepted and processed
// independently.
type BulkUpdateItemExecution struct {
	Index    int
	ID       uuid.UUID
	Revision uint64
	Phases   []BulkUpdatePhaseResult
}

// BulkUpdateExecution deepens the bulk Interface without changing the legacy
// HTTP-facing result. The application layer consumes phase outcomes; transport
// continues to consume Result.
type BulkUpdateExecution struct {
	Result *domainFacility.BulkOperationResult
	Items  []BulkUpdateItemExecution
}
