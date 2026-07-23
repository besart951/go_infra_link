package historycapture

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

type orderedObjectDataBacnetStore struct {
	domainObjectData.BacnetObjectStore
	created      bool
	linked       bool
	objectDataID uuid.UUID
	objectID     uuid.UUID
	err          error
}

func (s *orderedObjectDataBacnetStore) CreateForObjectData(
	_ context.Context,
	objectDataID uuid.UUID,
	entity *domainFacility.BacnetObject,
) error {
	if s.err != nil {
		return s.err
	}
	s.created = true
	s.linked = true
	s.objectDataID = objectDataID
	s.objectID = entity.ID
	return nil
}

type orderedObjectDataCreateChangeStore struct {
	ChangeStore
	persistence *orderedObjectDataBacnetStore
	recorded    bool
	table       string
	objectID    uuid.UUID
	err         error
}

func (s *orderedObjectDataCreateChangeStore) RecordCreate(
	_ context.Context,
	table string,
	objectID uuid.UUID,
) error {
	if !s.persistence.created || !s.persistence.linked {
		return errors.New("history recorded before ObjectData owner link")
	}
	s.recorded = true
	s.table = table
	s.objectID = objectID
	return s.err
}

func TestBacnetObjectCreateForObjectDataPersistsOwnerBeforeHistory(t *testing.T) {
	objectDataID := uuid.New()
	objectID := uuid.New()
	persistence := &orderedObjectDataBacnetStore{}
	changes := &orderedObjectDataCreateChangeStore{persistence: persistence}
	wrapper := WrapBacnetObject(persistence, changes)
	creator, ok := wrapper.(domainObjectData.ObjectDataBacnetObjectCreator)
	if !ok {
		t.Fatalf("wrapped store does not expose ObjectData create capability: %T", wrapper)
	}

	err := creator.CreateForObjectData(
		context.Background(),
		objectDataID,
		&domainFacility.BacnetObject{Base: domain.Base{ID: objectID}},
	)
	if err != nil {
		t.Fatalf("create for ObjectData: %v", err)
	}
	if persistence.objectDataID != objectDataID || persistence.objectID != objectID {
		t.Fatalf("persisted owner/object: got %s/%s, want %s/%s",
			persistence.objectDataID,
			persistence.objectID,
			objectDataID,
			objectID,
		)
	}
	if !changes.recorded || changes.table != "bacnet_objects" || changes.objectID != objectID {
		t.Fatalf("history record: recorded=%t table=%q id=%s", changes.recorded, changes.table, changes.objectID)
	}
}

func TestBacnetObjectCreateForObjectDataDoesNotRecordFailedPersistence(t *testing.T) {
	persistenceErr := errors.New("link failed")
	persistence := &orderedObjectDataBacnetStore{err: persistenceErr}
	changes := &orderedObjectDataCreateChangeStore{persistence: persistence}
	wrapper := WrapBacnetObject(persistence, changes)
	creator := wrapper.(domainObjectData.ObjectDataBacnetObjectCreator)

	err := creator.CreateForObjectData(
		context.Background(),
		uuid.New(),
		&domainFacility.BacnetObject{Base: domain.Base{ID: uuid.New()}},
	)
	if !errors.Is(err, persistenceErr) {
		t.Fatalf("error: got %v, want %v", err, persistenceErr)
	}
	if changes.recorded {
		t.Fatal("history recorded after failed persistence")
	}
}

var _ ChangeStore = (*orderedObjectDataCreateChangeStore)(nil)
var _ domainObjectData.ObjectDataBacnetObjectCreator = (*orderedObjectDataBacnetStore)(nil)
