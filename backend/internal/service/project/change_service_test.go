package project

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type changeStoreFake struct {
	changes []domainProject.NewChange
}

func TestChangeSemanticsUsesClosedVocabulary(t *testing.T) {
	tests := []struct {
		event     string
		aggregate domainProject.AggregateKind
		action    domainProject.ChangeAction
	}{
		{"project.user.invited", domainProject.AggregateProjectUser, domainProject.ChangeInvited},
		{"project.phase.changed", domainProject.AggregateProject, domainProject.ChangeUpdated},
		{"project.field_device.multi_created", domainProject.AggregateFieldDevice, domainProject.ChangeCreated},
		{"project.sps_controller.ip_address.changed", domainProject.AggregateSPSController, domainProject.ChangeUpdated},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			got, err := changeSemantics(tt.event)
			if err != nil {
				t.Fatal(err)
			}
			if got.aggregate != tt.aggregate || got.action != tt.action {
				t.Fatalf("semantics = %q/%q, want %q/%q", got.aggregate, got.action, tt.aggregate, tt.action)
			}
		})
	}
}

func TestChangeServiceRejectsUnknownEvent(t *testing.T) {
	service := NewChangeService(&changeStoreFake{})
	err := service.RecordEvent(t.Context(), uuid.New(), "project.user.promoted", nil)
	if !errors.Is(err, ErrUnknownProjectEvent) {
		t.Fatalf("error = %v, want %v", err, ErrUnknownProjectEvent)
	}
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
	}
}

func TestChangeServiceUsesExplicitChangedFieldsWhenProvided(t *testing.T) {
	store := &changeStoreFake{}
	service := NewChangeService(store)
	projectID, fieldDeviceID := uuid.New(), uuid.New()

	_, err := service.RecordEventsWithFields(context.Background(), projectID, "project.field_device.updated", nil, []string{"description"}, fieldDeviceID.String())
	if err != nil {
		t.Fatalf("record event: %v", err)
	}
	if len(store.changes) != 1 {
		t.Fatalf("recorded %d changes, want 1", len(store.changes))
	}
	if got, want := store.changes[0].ChangedFields, []string{"description"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("changed fields = %#v, want %#v", got, want)
	}
}
