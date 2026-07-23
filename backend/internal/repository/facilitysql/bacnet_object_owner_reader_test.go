package facilitysql

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
)

func TestBacnetObjectOwnerReaderReturnsOnlyRequestedPersistedOwners(t *testing.T) {
	db := newBacnetReferenceUsageRepoTestDB(t)
	requestedID := uuid.New()
	unrelatedID := uuid.New()
	projectOne := uuid.New()
	projectTwo := uuid.New()
	objectDataOne := uuid.New()
	objectDataTwo := uuid.New()
	globalObjectData := uuid.New()

	for _, item := range []domainFacility.ObjectData{
		{Base: domain.Base{ID: objectDataOne}, Description: "One", Version: "1", ProjectID: &projectOne},
		{Base: domain.Base{ID: objectDataTwo}, Description: "Two", Version: "1", ProjectID: &projectTwo},
		{Base: domain.Base{ID: globalObjectData}, Description: "Global", Version: "1"},
	} {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed ObjectData %s: %v", item.ID, err)
		}
	}
	for _, link := range []struct {
		bacnetObjectID uuid.UUID
		objectDataID   uuid.UUID
	}{
		{requestedID, objectDataTwo},
		{requestedID, objectDataOne},
		{requestedID, globalObjectData},
		{unrelatedID, objectDataOne},
	} {
		if err := db.Table("object_data_bacnet_objects").Create(map[string]any{
			"bacnet_object_id": link.bacnetObjectID,
			"object_data_id":   link.objectDataID,
		}).Error; err != nil {
			t.Fatalf("seed ObjectData BACnet link: %v", err)
		}
	}

	owners, err := NewBacnetObjectOwnerReader(db).GetByBacnetObjectIDs(
		context.Background(),
		[]uuid.UUID{requestedID},
	)
	if err != nil {
		t.Fatalf("resolve owners: %v", err)
	}
	want := []domainObjectData.BacnetObjectOwner{
		{BacnetObjectID: requestedID, ObjectDataID: objectDataOne, ProjectID: &projectOne},
		{BacnetObjectID: requestedID, ObjectDataID: objectDataTwo, ProjectID: &projectTwo},
		{BacnetObjectID: requestedID, ObjectDataID: globalObjectData},
	}
	// The reader sorts by UUID, so sort the expected projection through the same
	// stable relationship ordering before comparing values.
	sortBacnetObjectOwnersForTest(want)
	if !reflect.DeepEqual(owners, want) {
		t.Fatalf("owners: got %+v, want %+v", owners, want)
	}
}

func TestBacnetObjectOwnerReaderSkipsDatabaseForEmptyInput(t *testing.T) {
	owners, err := NewBacnetObjectOwnerReader(nil).GetByBacnetObjectIDs(
		context.Background(),
		nil,
	)
	if err != nil || len(owners) != 0 {
		t.Fatalf("empty owners: got %+v, err=%v", owners, err)
	}
}

func sortBacnetObjectOwnersForTest(owners []domainObjectData.BacnetObjectOwner) {
	sort.Slice(owners, func(i, j int) bool {
		return owners[i].ObjectDataID.String() < owners[j].ObjectDataID.String()
	})
}
