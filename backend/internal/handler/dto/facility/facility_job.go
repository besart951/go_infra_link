package facility

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FacilityJobResponse is the persisted representation of every asynchronous
// Facility operation.
type FacilityJobResponse struct {
	JobID        uuid.UUID       `json:"job_id"`
	Kind         string          `json:"kind" enums:"control_cabinet,sps_controller,sps_controller_system_type,field_device,object_data"`
	Type         string          `json:"type" enums:"copy,export,bulk,delete,restore"`
	Class        string          `json:"class" enums:"mutation,export"`
	Status       string          `json:"status" enums:"queued,running,completed,failed"`
	Progress     int             `json:"progress" minimum:"0" maximum:"100"`
	Stage        string          `json:"stage"`
	Error        string          `json:"error,omitempty"`
	Attempts     int             `json:"attempts"`
	Processed    int64           `json:"processed"`
	Total        *int64          `json:"total,omitempty"`
	SuccessCount int64           `json:"success_count"`
	FailureCount int64           `json:"failure_count"`
	Retryable    bool            `json:"retryable"`
	Result       json.RawMessage `json:"result,omitempty" swaggertype:"object"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

type FacilityJobListResponse struct {
	Items          []FacilityJobResponse `json:"items"`
	NextCursor     string                `json:"next_cursor,omitempty"`
	PreviousCursor string                `json:"previous_cursor,omitempty"`
}
