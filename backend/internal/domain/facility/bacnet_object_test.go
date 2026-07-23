package facility

import (
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

func TestBacnetObjectApplyPatchAndReassign(t *testing.T) {
	objectID := uuid.New()
	oldFieldDeviceID := uuid.New()
	newFieldDeviceID := uuid.New()
	textFix := "new"
	visible := true
	softwareType := BacnetSoftwareTypeAO
	object := &BacnetObject{
		Base:          domain.Base{ID: objectID},
		TextFix:       "old",
		SoftwareType:  BacnetSoftwareTypeAI,
		FieldDeviceID: &oldFieldDeviceID,
	}

	err := object.ApplyPatch(BacnetObjectPatch{
		ID:           objectID,
		TextFix:      &textFix,
		GMSVisible:   &visible,
		SoftwareType: &softwareType,
	})
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}
	if err := object.AssignToFieldDevice(newFieldDeviceID); err != nil {
		t.Fatalf("reassign: %v", err)
	}
	if object.TextFix != textFix || !object.GMSVisible || object.SoftwareType != softwareType ||
		object.FieldDeviceID == nil || *object.FieldDeviceID != newFieldDeviceID {
		t.Fatalf("unexpected object state: %+v", object)
	}

	if err := object.DetachFromFieldDevice(); err != nil {
		t.Fatalf("detach: %v", err)
	}
	if object.FieldDeviceID != nil {
		t.Fatalf("expected detached object, got owner %v", object.FieldDeviceID)
	}
}

func TestBacnetObjectRejectsMismatchedPatchAndInvalidOwner(t *testing.T) {
	object := &BacnetObject{Base: domain.Base{ID: uuid.New()}}
	if err := object.ApplyPatch(BacnetObjectPatch{ID: uuid.New()}); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("mismatched patch: got %v", err)
	}
	if err := object.AssignToFieldDevice(uuid.Nil); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("nil owner: got %v", err)
	}
}
