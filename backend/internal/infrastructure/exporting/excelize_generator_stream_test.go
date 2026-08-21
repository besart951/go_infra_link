package exporting

import (
	"context"
	"fmt"
	"testing"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type generatorPageSource struct {
	items []domainFacility.FieldDevice
	calls int
}

func (s *generatorPageSource) ResolveControllers(context.Context, domainExport.Request) ([]domainExport.Controller, error) {
	return nil, nil
}

func (s *generatorPageSource) ListFieldDevicesByControllerAfter(_ context.Context, _ uuid.UUID, _ domainExport.Request, afterID uuid.UUID, limit int) ([]domainFacility.FieldDevice, error) {
	s.calls++
	start := 0
	if afterID != uuid.Nil {
		for i := range s.items {
			if s.items[i].ID == afterID {
				start = i + 1
				break
			}
		}
	}
	if start >= len(s.items) {
		return []domainFacility.FieldDevice{}, nil
	}
	end := min(start+limit, len(s.items))
	return append([]domainFacility.FieldDevice(nil), s.items[start:end]...), nil
}

func TestGenerateWorkbookStreamsPastFormerThreeHundredRowLimit(t *testing.T) {
	controllerID := uuid.New()
	items := make([]domainFacility.FieldDevice, 301)
	for i := range items {
		bmk := fmt.Sprintf("FD-%03d", i+1)
		items[i] = domainFacility.FieldDevice{BMK: &bmk, ApparatNr: i + 1}
		items[i].ID = uuid.New()
		items[i].SystemPart.Name = "Part"
		items[i].Apparat.Name = "Device"
	}
	source := &generatorPageSource{items: items}
	path := t.TempDir() + "/export.xlsx"
	count, err := NewExcelizeGenerator().GenerateWorkbook(t.Context(), path, []domainExport.Controller{{
		ID: controllerID, ControlCabinetID: uuid.New(), GADevice: "A",
	}}, source, domainExport.Request{SchemaVersion: 1}, 300)
	if err != nil {
		t.Fatalf("GenerateWorkbook() error = %v", err)
	}
	if count != 301 || source.calls != 2 {
		t.Fatalf("count=%d calls=%d, want 301 devices in two keyset pages", count, source.calls)
	}
	workbook, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer func() { _ = workbook.Close() }()
	for _, sheet := range []string{"Export-Manifest", "Data-FieldDevices", "Data-Specifications", "Data-BACnetObjects", "Data-AlarmValues"} {
		if index, err := workbook.GetSheetIndex(sheet); err != nil || index < 0 {
			t.Fatalf("missing machine sheet %q (index=%d, error=%v)", sheet, index, err)
		}
	}
}
