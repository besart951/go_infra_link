package importing

import (
	"bytes"
	"context"
	"testing"
	"time"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type captureSink struct {
	devices    []domainFacility.FieldDevice
	specs      []domainFacility.Specification
	objects    []domainFacility.BacnetObject
	references []fielddeviceimport.SoftwareReference
	alarms     []domainFacility.BacnetObjectAlarmValue
}

func (s *captureSink) FieldDevices(_ context.Context, values []domainFacility.FieldDevice) error {
	s.devices = append(s.devices, values...)
	return nil
}
func (s *captureSink) Specifications(_ context.Context, values []domainFacility.Specification) error {
	s.specs = append(s.specs, values...)
	return nil
}
func (s *captureSink) BacnetObjects(_ context.Context, values []domainFacility.BacnetObject) error {
	s.objects = append(s.objects, values...)
	return nil
}
func (s *captureSink) SoftwareReferences(_ context.Context, values []fielddeviceimport.SoftwareReference) error {
	s.references = append(s.references, values...)
	return nil
}
func (s *captureSink) AlarmValues(_ context.Context, values []domainFacility.BacnetObjectAlarmValue) error {
	s.alarms = append(s.alarms, values...)
	return nil
}

func TestExcelizeReaderReadsVersionedDataTabs(t *testing.T) {
	fixture := newImportFixture(t)
	sink := &captureSink{}

	manifest, err := NewExcelizeReader().Read(context.Background(), bytes.NewReader(fixture), sink)

	if err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 2 || manifest.DeviceCount != 1 {
		t.Fatalf("manifest = %+v", manifest)
	}
	if len(sink.devices) != 1 || len(sink.specs) != 1 || len(sink.objects) != 2 || len(sink.references) != 1 || len(sink.alarms) != 1 {
		t.Fatalf("unexpected rows: devices=%d specs=%d objects=%d refs=%d alarms=%d", len(sink.devices), len(sink.specs), len(sink.objects), len(sink.references), len(sink.alarms))
	}
	if sink.objects[0].SoftwareReferenceID == nil || *sink.objects[0].SoftwareReferenceID != sink.objects[1].ID {
		t.Fatalf("software reference was not preserved: %+v", sink.objects[0])
	}
}

func newImportFixture(t *testing.T) []byte {
	return newImportFixtureForDevice(t, uuid.New(), 1)
}

func newImportFixtureForDevice(t *testing.T, deviceID uuid.UUID, deviceCount int) []byte {
	t.Helper()
	file := excelize.NewFile()
	defer file.Close()
	file.DeleteSheet("Sheet1")
	objectID, targetID := uuid.New(), uuid.New()
	assignmentID, systemPartID, apparatID := uuid.New(), uuid.New(), uuid.New()
	now := time.Unix(1_700_000_000, 0).UTC().Format(time.RFC3339Nano)
	addSheet(t, file, "Export-Manifest", [][]any{
		{"schema_version", 2}, {"snapshot_at", now}, {"scope", "global"}, {"device_count", deviceCount},
		{"specification_count", 1}, {"bacnet_object_count", 2}, {"software_reference_count", 1}, {"alarm_value_count", 1},
	})
	addSheet(t, file, "Data-FieldDevices", [][]any{
		{"source_id", "version", "created_at", "updated_at", "sps_controller_id", "sps_controller_system_type_id", "system_part_id", "apparat_id", "apparat_nr", "bmk", "description", "text_individual"},
		{deviceID, 3, now, now, uuid.New(), assignmentID, systemPartID, apparatID, 7, "BMK", "Device", "Custom"},
	})
	addSheet(t, file, "Data-Specifications", [][]any{
		{"source_id", "field_device_id", "version", "supplier", "brand", "type", "motor_valve", "size", "installation_location", "ph", "acdc", "amperage", "power", "rotation"},
		{uuid.New(), deviceID, 2, "Supplier", "Brand", "Type", "", 10, "Room", 3, "AC", 2.5, 10.2, 100},
	})
	addSheet(t, file, "Data-BACnetObjects", [][]any{
		{"source_id", "field_device_id", "version", "text_fix", "description", "gms_visible", "optional", "text_individual", "software_type", "software_number", "hardware_type", "hardware_quantity", "software_reference_id", "state_text_id", "notification_class_id", "alarm_type_id", "alarm_definition_id"},
		{objectID, deviceID, 1, "Object", "Description", true, false, "", "ai", 1, "ai", 1, targetID, "", "", "", ""},
		{targetID, deviceID, 1, "Target", "", true, true, "", "av", 2, "", 0, "", "", "", "", ""},
	})
	addSheet(t, file, "Data-SoftwareReferences", [][]any{
		{"source_object_id", "target_object_id", "field_device_id"}, {objectID, targetID, deviceID},
	})
	addSheet(t, file, "Data-AlarmValues", [][]any{
		{"source_id", "bacnet_object_id", "version", "alarm_type_field_id", "value_number", "value_integer", "value_boolean", "value_string", "value_json", "unit_id", "source"},
		{uuid.New(), objectID, 1, uuid.New(), 12.5, "", "", "", "", "", "import"},
	})
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func addSheet(t *testing.T, file *excelize.File, name string, rows [][]any) {
	t.Helper()
	if _, err := file.NewSheet(name); err != nil {
		t.Fatal(err)
	}
	for index, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, index+1)
		if err := file.SetSheetRow(name, cell, &row); err != nil {
			t.Fatal(err)
		}
	}
}
