package historysql

import (
	"context"
	"testing"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

func TestHasHistoricalProjectControlCabinetRequiresRootLinkEventAndBothScopes(t *testing.T) {
	db := newHistoryStoreTestDB(t)
	store := NewStore(db)
	controlCabinetID := uuid.New()
	projectID := uuid.New()
	otherProjectID := uuid.New()
	eventID := uuid.New()
	otherEventID := uuid.New()
	now := time.Now().UTC()
	events := []domainHistory.ChangeEvent{
		{
			ID:          eventID,
			OccurredAt:  now,
			Action:      domainHistory.ActionDelete,
			EntityTable: "project_control_cabinets",
			EntityID:    uuid.New(),
		},
		{
			ID:          otherEventID,
			OccurredAt:  now,
			Action:      domainHistory.ActionUpdate,
			EntityTable: "field_devices",
			EntityID:    uuid.New(),
		},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("create events: %v", err)
	}
	scopes := []domainHistory.ChangeEventScope{
		{
			ID:            uuid.New(),
			ChangeEventID: eventID,
			ScopeType:     scopeControlCabinet,
			ScopeID:       controlCabinetID,
			OccurredAt:    now,
		},
		{
			ID:            uuid.New(),
			ChangeEventID: eventID,
			ScopeType:     scopeProject,
			ScopeID:       projectID,
			OccurredAt:    now,
		},
		{
			ID:            uuid.New(),
			ChangeEventID: otherEventID,
			ScopeType:     scopeControlCabinet,
			ScopeID:       controlCabinetID,
			OccurredAt:    now,
		},
		{
			ID:            uuid.New(),
			ChangeEventID: otherEventID,
			ScopeType:     scopeProject,
			ScopeID:       otherProjectID,
			OccurredAt:    now,
		},
	}
	if err := db.Create(&scopes).Error; err != nil {
		t.Fatalf("create scopes: %v", err)
	}

	hasScope, err := store.HasHistoricalProjectControlCabinet(
		context.Background(),
		projectID,
		controlCabinetID,
	)
	if err != nil {
		t.Fatalf("HasHistoricalProjectControlCabinet returned error: %v", err)
	}
	if !hasScope {
		t.Fatal("expected matching project and cabinet scopes")
	}

	for _, test := range []struct {
		name             string
		controlCabinetID uuid.UUID
		projectID        uuid.UUID
	}{
		{name: "descendant-only project event", controlCabinetID: controlCabinetID, projectID: otherProjectID},
		{name: "different cabinet", controlCabinetID: uuid.New(), projectID: projectID},
		{name: "zero cabinet", projectID: projectID},
		{name: "zero project", controlCabinetID: controlCabinetID},
	} {
		t.Run(test.name, func(t *testing.T) {
			hasScope, err := store.HasHistoricalProjectControlCabinet(
				context.Background(),
				test.projectID,
				test.controlCabinetID,
			)
			if err != nil {
				t.Fatalf("HasHistoricalProjectControlCabinet returned error: %v", err)
			}
			if hasScope {
				t.Fatal("unexpected historical association")
			}
		})
	}
}
