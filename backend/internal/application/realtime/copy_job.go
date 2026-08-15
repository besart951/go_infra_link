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
	Status    string
	Progress  int
	Stage     string
	Error     string
	UpdatedAt time.Time
}

// CopyJobProgressPublisher delivers a copy job's latest state to the owning
// user's existing realtime connection.
type CopyJobProgressPublisher interface {
	BroadcastCopyJobProgress(context.Context, CopyJobProgressEvent)
}
