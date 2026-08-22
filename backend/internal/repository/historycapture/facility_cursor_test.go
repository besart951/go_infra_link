package historycapture

import (
	"context"
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
)

type cursorFieldDeviceStoreStub struct {
	domainFieldDevice.FieldDeviceStore
	page *domainFacility.FieldDeviceCursorPage
}

func (s cursorFieldDeviceStoreStub) GetCursorPage(_ context.Context, _ domainFacility.FieldDeviceCursorQuery) (*domainFacility.FieldDeviceCursorPage, error) {
	return s.page, nil
}

func TestFieldDeviceStorePreservesCursorReader(t *testing.T) {
	wrapper := WrapFieldDevice(cursorFieldDeviceStoreStub{
		page: &domainFacility.FieldDeviceCursorPage{},
	}, nil)

	reader, ok := wrapper.(domainFieldDevice.CursorReader)
	if !ok {
		t.Fatal("history decorator must preserve the cursor reader capability")
	}
	if _, err := reader.GetCursorPage(t.Context(), domainFacility.FieldDeviceCursorQuery{}); err != nil {
		t.Fatalf("cursor page: %v", err)
	}
}
