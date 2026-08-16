package project

import (
	"context"
	"strings"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ChangeService struct {
	store domainProject.ChangeStore
}

func NewChangeService(store domainProject.ChangeStore) *ChangeService {
	return &ChangeService{store: store}
}

func (s *ChangeService) ListAfter(ctx context.Context, projectID uuid.UUID, afterRevision uint64, limit int) (*domainProject.ChangePage, error) {
	return s.store.ListAfter(ctx, projectID, afterRevision, limit)
}

// RecordEvent converts the existing semantic project event vocabulary into one
// durable aggregate event per entity. It is deliberately independent of the
// WebSocket collaboration payload.
func (s *ChangeService) RecordEvent(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, entityIDs ...string) error {
	_, err := s.RecordEvents(ctx, projectID, eventType, actorID, entityIDs...)
	return err
}

// RecordEvents persists semantic changes and returns their allocated project
// revisions so the realtime transport can publish the durable cursor exactly.
func (s *ChangeService) RecordEvents(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, entityIDs ...string) ([]domainProject.Change, error) {
	return s.RecordEventsWithFields(ctx, projectID, eventType, actorID, nil, entityIDs...)
}

// RecordEventsWithFields records one durable project change per entity. A nil
// changedFields value preserves the legacy semantic defaults; a non-nil value
// is the precise set supplied by the mutation request contract.
func (s *ChangeService) RecordEventsWithFields(ctx context.Context, projectID uuid.UUID, eventType string, actorID *uuid.UUID, changedFields []string, entityIDs ...string) ([]domainProject.Change, error) {
	aggregateType, action := changeSemantics(eventType)
	if changedFields == nil {
		changedFields = semanticChangedFields(aggregateType, action)
	}
	ids := parseEntityIDs(entityIDs)
	inputs := make([]domainProject.NewChange, 0, max(1, len(ids)))
	if len(ids) == 0 {
		if aggregateType == "project" {
			ids = append(ids, projectID)
		} else {
			inputs = append(inputs, newChange(projectID, aggregateType, nil, action, actorID, changedFields))
		}
	}
	for i := range ids {
		id := ids[i]
		inputs = append(inputs, newChange(projectID, aggregateType, &id, action, actorID, changedFields))
	}
	if batch, ok := s.store.(domainProject.BatchChangeStore); ok {
		return batch.AppendBatch(ctx, inputs)
	}
	changes := make([]domainProject.Change, 0, len(inputs))
	for _, input := range inputs {
		change, err := s.store.Append(ctx, input)
		if err != nil {
			return nil, err
		}
		changes = append(changes, *change)
	}
	return changes, nil
}

func newChange(projectID uuid.UUID, aggregateType string, aggregateID *uuid.UUID, action domainProject.ChangeAction, actorID *uuid.UUID, changedFields []string) domainProject.NewChange {
	return domainProject.NewChange{
		ProjectID:     projectID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Action:        action,
		ActorID:       actorID,
		ChangedFields: changedFields,
		ParentRefs:    map[string]uuid.UUID{"project_id": projectID},
	}
}

func semanticChangedFields(aggregateType string, action domainProject.ChangeAction) []string {
	if action != domainProject.ChangeUpdated {
		return []string{}
	}
	switch aggregateType {
	case "project":
		return []string{"name", "description", "status", "start_date", "phase_id"}
	case "control_cabinet":
		return []string{"building_id", "control_cabinet_nr"}
	case "sps_controller":
		return []string{"control_cabinet_id", "ga_device", "device_name", "device_description", "device_location", "ip_address", "subnet", "gateway", "vlan", "system_types"}
	case "sps_controller_system_type":
		return []string{"number", "document_name"}
	case "field_device":
		return []string{"bmk", "description", "text_fix", "apparat_nr", "sps_controller_system_type_id", "system_part_id", "apparat_id", "specification", "bacnet_objects"}
	default:
		return []string{}
	}
}

func changeSemantics(eventType string) (string, domainProject.ChangeAction) {
	parts := strings.Split(strings.TrimSpace(eventType), ".")
	if len(parts) < 2 || parts[0] != "project" {
		return "project", domainProject.ChangeAction("changed")
	}
	action := parts[len(parts)-1]
	if action == "multi_created" {
		action = string(domainProject.ChangeCreated)
	}
	aggregateType := strings.Join(parts[1:len(parts)-1], "_")
	if aggregateType == "" {
		aggregateType = "project"
	}
	if aggregateType == "phase" {
		aggregateType = "project"
	}
	return aggregateType, domainProject.ChangeAction(action)
}

func parseEntityIDs(values []string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
