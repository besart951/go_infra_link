package fielddevice

import (
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestToFieldDeviceResponseUsesSharedScalarContract(t *testing.T) {
	now := time.Now().UTC()
	fieldDeviceID := uuid.New()
	systemPartID := uuid.New()

	response := toFieldDeviceResponse(&domainFacility.FieldDevice{
		Base:         domain.Base{ID: fieldDeviceID, Version: 3, CreatedAt: now, UpdatedAt: now},
		SystemPartID: systemPartID,
	})

	if response == nil {
		t.Fatal("expected response")
	}
	if response.ID != fieldDeviceID || response.Version != 3 {
		t.Fatalf("expected id/version %s/3, got %s/%d", fieldDeviceID, response.ID, response.Version)
	}
	if response.SystemPartID == nil || *response.SystemPartID != systemPartID {
		t.Fatalf("expected system part %s, got %#v", systemPartID, response.SystemPartID)
	}
}
