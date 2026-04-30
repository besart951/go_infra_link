package changecapture

import (
	"context"

	"github.com/google/uuid"
)

type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

type EntityRef struct {
	Domain string
	Type   string
	ID     uuid.UUID
}

type Change struct {
	Action   Action
	Entity   EntityRef
	Metadata map[string]string
}

type Recorder interface {
	Record(ctx context.Context, change Change) error
}

type NoopRecorder struct{}

func (NoopRecorder) Record(context.Context, Change) error {
	return nil
}

func DefaultRecorder(recorder Recorder) Recorder {
	if recorder == nil {
		return NoopRecorder{}
	}
	return recorder
}
