package realtime

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CopyJobProgressEvent is the user-scoped realtime representation of a
// long-running facility copy. OwnerID is transport metadata and is never sent
// to browser clients.
type CopyJobProgressEvent struct {
	JobID     uuid.UUID
	OwnerID   uuid.UUID
	Kind      string
	JobType   string
	Class     string
	Status    string
	Progress  int
	Stage     string
	Error     string
	Processed int64
	Total     *int64
	Succeeded int64
	Failed    int64
	UpdatedAt time.Time
}

// CopyJobProgressPublisher delivers a copy job's latest state to the owning
// user's existing realtime connection.
type CopyJobProgressPublisher interface {
	BroadcastCopyJobProgress(context.Context, CopyJobProgressEvent)
}
