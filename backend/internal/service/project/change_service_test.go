package project

import (
	"context"
	"testing"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type changeStoreFake struct {
	changes []domainProject.NewChange
}

func (f *changeStoreFake) Append(_ context.Context, change domainProject.NewChange) (*domainProject.Change, error) {
	f.changes = append(f.changes, change)
	return &domainProject.Change{}, nil
}

func (f *changeStoreFake) ListAfter(context.Context, uuid.UUID, uint64, int) (*domainProject.ChangePage, error) {
	return nil, nil
}

func TestChangeServiceRecordsOneSemanticEventPerAggregate(t *testing.T) {
	store := &changeStoreFake{}
	service := NewChangeService(store)
	projectID, firstID, secondID := uuid.New(), uuid.New(), uuid.New()

	err := service.RecordEvent(context.Background(), projectID, "project.field_device.multi_created", nil, firstID.String(), secondID.String())
	if err != nil {
		t.Fatalf("record event: %v", err)
	}
	if len(store.changes) != 2 {
		t.Fatalf("recorded %d changes, want 2", len(store.changes))
	}
	for _, change := range store.changes {
		if change.AggregateType != "field_device" || change.Action != domainProject.ChangeCreated {
			t.Fatalf("unexpected semantic change: %+v", change)
		}
		if change.ParentRefs["project_id"] != projectID {
			t.Fatalf("missing project parent ref: %+v", change.ParentRefs)
		}
	}
}
