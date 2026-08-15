package facility

import (
	"time"

	"github.com/google/uuid"
)

// CopyJobResponse is returned immediately after starting a hierarchy copy and
// is also used to reconcile a job after a WebSocket reconnect.
type CopyJobResponse struct {
	JobID     uuid.UUID `json:"job_id"`
	Kind      string    `json:"kind" enums:"control_cabinet,sps_controller,sps_controller_system_type"`
	Status    string    `json:"status" enums:"queued,running,completed,failed"`
	Progress  int       `json:"progress" minimum:"0" maximum:"100"`
	Stage     string    `json:"stage"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
