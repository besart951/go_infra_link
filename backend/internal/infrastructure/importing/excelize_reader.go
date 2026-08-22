package importing

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	fielddeviceimport "github.com/besart951/go_infra_link/backend/internal/application/fielddeviceimport"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

const importBatchSize = 500

type ExcelizeReader struct{}

func NewExcelizeReader() ExcelizeReader { return ExcelizeReader{} }

func (ExcelizeReader) Read(ctx context.Context, source io.Reader, sink fielddeviceimport.Sink) (fielddeviceimport.Manifest, error) {
	workbook, err := excelize.OpenReader(source)
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	defer workbook.Close()
	manifest, err := readManifest(workbook)
	if err != nil {
		return manifest, err
	}
	for _, scan := range workbookScanners(workbook, sink) {
		if err := scan(ctx); err != nil {
			return manifest, err
		}
	}
	return manifest, nil
}

type workbookScan func(context.Context) error

func workbookScanners(file *excelize.File, sink fielddeviceimport.Sink) []workbookScan {
	return []workbookScan{
		func(ctx context.Context) error { return scanFieldDevices(ctx, file, sink) },
		func(ctx context.Context) error { return scanSpecifications(ctx, file, sink) },
		func(ctx context.Context) error { return scanBacnetObjects(ctx, file, sink) },
		func(ctx context.Context) error { return scanSoftwareReferences(ctx, file, sink) },
		func(ctx context.Context) error { return scanAlarmValues(ctx, file, sink) },
	}
}

func readManifest(file *excelize.File) (fielddeviceimport.Manifest, error) {
	values := make(map[string]string)
	rows, err := file.Rows("Export-Manifest")
	if err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	defer rows.Close()
	for rows.Next() {
		columns, rowErr := rows.Columns()
		if rowErr != nil {
			return fielddeviceimport.Manifest{}, rowErr
		}
		row := rowValues{columns: columns}
		values[row.at(0)] = row.at(1)
	}
	if err := rows.Error(); err != nil {
		return fielddeviceimport.Manifest{}, err
	}
	version, err := strconv.Atoi(values["schema_version"])
	if err != nil {
		return fielddeviceimport.Manifest{}, fmt.Errorf("invalid manifest schema_version: %w", err)
	}
	count, err := strconv.ParseInt(values["device_count"], 10, 64)
	if err != nil {
		return fielddeviceimport.Manifest{}, fmt.Errorf("invalid manifest device_count: %w", err)
	}
	snapshot, err := time.Parse(time.RFC3339Nano, values["snapshot_at"])
	if err != nil {
		return fielddeviceimport.Manifest{}, fmt.Errorf("invalid manifest snapshot_at: %w", err)
	}
	counts, err := parseManifestCounts(values)
	return fielddeviceimport.Manifest{SchemaVersion: version, SnapshotAt: snapshot, Scope: values["scope"], DeviceCount: count, Counts: counts}, err
}

func parseManifestCounts(values map[string]string) (fielddeviceimport.Counts, error) {
	keys := []string{"specification_count", "bacnet_object_count", "software_reference_count", "alarm_value_count"}
	counts := make([]int64, len(keys))
	for index, key := range keys {
		value, err := strconv.ParseInt(values[key], 10, 64)
		if err != nil {
			return fielddeviceimport.Counts{}, fmt.Errorf("invalid manifest %s: %w", key, err)
		}
		counts[index] = value
	}
	return fielddeviceimport.Counts{
		Specifications: counts[0], BacnetObjects: counts[1],
		SoftwareReferences: counts[2], AlarmValues: counts[3],
	}, nil
}

type rowValues struct {
	columns []string
	index   map[string]int
}

func (r rowValues) at(index int) string {
	if index < 0 || index >= len(r.columns) {
		return ""
	}
	return strings.TrimSpace(r.columns[index])
}

func (r rowValues) get(name string) string {
	index, ok := r.index[name]
	if !ok {
		return ""
	}
	return r.at(index)
}

func scanRows(file *excelize.File, sheet string, consume func(rowValues) error) error {
	rows, err := file.Rows(sheet)
	if err != nil {
		return fmt.Errorf("open sheet %s: %w", sheet, err)
	}
	defer rows.Close()
	if !rows.Next() {
		return fmt.Errorf("sheet %s is empty", sheet)
	}
	headings, err := rows.Columns()
	if err != nil {
		return err
	}
	index := headingIndex(headings)
	rowNumber := 1
	for rows.Next() {
		rowNumber++
		columns, rowErr := rows.Columns()
		if rowErr != nil {
			return rowErr
		}
		if err := consume(rowValues{columns: columns, index: index}); err != nil {
			return fmt.Errorf("sheet %s row %d: %w", sheet, rowNumber, err)
		}
	}
	return rows.Error()
}

func headingIndex(headings []string) map[string]int {
	index := make(map[string]int, len(headings))
	for position, heading := range headings {
		index[strings.TrimSpace(heading)] = position
	}
	return index
}

func scanFieldDevices(ctx context.Context, file *excelize.File, sink fielddeviceimport.Sink) error {
	batch := make([]domainFacility.FieldDevice, 0, importBatchSize)
	err := scanRows(file, "Data-FieldDevices", func(row rowValues) error {
		value, err := parseFieldDevice(row)
		if err != nil {
			return err
		}
		batch = append(batch, value)
		if len(batch) < importBatchSize {
			return nil
		}
		err = sink.FieldDevices(ctx, batch)
		batch = make([]domainFacility.FieldDevice, 0, importBatchSize)
		return err
	})
	if err != nil || len(batch) == 0 {
		return err
	}
	return sink.FieldDevices(ctx, batch)
}

func parseFieldDevice(row rowValues) (domainFacility.FieldDevice, error) {
	base, err := parseBase(row)
	if err != nil {
		return domainFacility.FieldDevice{}, err
	}
	assignmentID, err := parseUUID(row.get("sps_controller_system_type_id"))
	if err != nil {
		return domainFacility.FieldDevice{}, err
	}
	systemPartID, err := parseUUID(row.get("system_part_id"))
	if err != nil {
		return domainFacility.FieldDevice{}, err
	}
	apparatID, err := parseUUID(row.get("apparat_id"))
	if err != nil {
		return domainFacility.FieldDevice{}, err
	}
	apparatNr, err := strconv.Atoi(row.get("apparat_nr"))
	return domainFacility.FieldDevice{
		Base: base, SPSControllerSystemTypeID: assignmentID, SystemPartID: systemPartID,
		ApparatID: apparatID, ApparatNr: apparatNr, BMK: optionalString(row.get("bmk")),
		Description: optionalString(row.get("description")), TextIndividuell: optionalString(row.get("text_individual")),
	}, err
}

func scanSpecifications(ctx context.Context, file *excelize.File, sink fielddeviceimport.Sink) error {
	batch := make([]domainFacility.Specification, 0, importBatchSize)
	err := scanRows(file, "Data-Specifications", func(row rowValues) error {
		value, err := parseSpecification(row)
		if err != nil {
			return err
		}
		batch = append(batch, value)
		if len(batch) == importBatchSize {
			err = sink.Specifications(ctx, batch)
			batch = make([]domainFacility.Specification, 0, importBatchSize)
		}
		return err
	})
	if err != nil || len(batch) == 0 {
		return err
	}
	return sink.Specifications(ctx, batch)
}

func parseSpecification(row rowValues) (domainFacility.Specification, error) {
	base, err := parseBase(row)
	if err != nil {
		return domainFacility.Specification{}, err
	}
	ownerID, err := parseUUID(row.get("field_device_id"))
	if err != nil {
		return domainFacility.Specification{}, err
	}
	return domainFacility.Specification{
		Base: base, FieldDeviceID: &ownerID,
		SpecificationSupplier: optionalString(row.get("supplier")), SpecificationBrand: optionalString(row.get("brand")),
		SpecificationType: optionalString(row.get("type")), AdditionalInfoMotorValve: optionalString(row.get("motor_valve")),
		AdditionalInfoSize: optionalInt(row.get("size")), AdditionalInformationInstallationLocation: optionalString(row.get("installation_location")),
		ElectricalConnectionPH: optionalInt(row.get("ph")), ElectricalConnectionACDC: optionalString(row.get("acdc")),
		ElectricalConnectionAmperage: optionalFloat(row.get("amperage")), ElectricalConnectionPower: optionalFloat(row.get("power")),
		ElectricalConnectionRotation: optionalInt(row.get("rotation")),
	}, nil
}

func scanBacnetObjects(ctx context.Context, file *excelize.File, sink fielddeviceimport.Sink) error {
	batch := make([]domainFacility.BacnetObject, 0, importBatchSize)
	err := scanRows(file, "Data-BACnetObjects", func(row rowValues) error {
		value, err := parseBacnetObject(row)
		if err != nil {
			return err
		}
		batch = append(batch, value)
		if len(batch) == importBatchSize {
			err = sink.BacnetObjects(ctx, batch)
			batch = make([]domainFacility.BacnetObject, 0, importBatchSize)
		}
		return err
	})
	if err != nil || len(batch) == 0 {
		return err
	}
	return sink.BacnetObjects(ctx, batch)
}

func parseBacnetObject(row rowValues) (domainFacility.BacnetObject, error) {
	base, err := parseBase(row)
	if err != nil {
		return domainFacility.BacnetObject{}, err
	}
	ownerID, err := parseUUID(row.get("field_device_id"))
	if err != nil {
		return domainFacility.BacnetObject{}, err
	}
	softwareNumber, err := parseUint16(row.get("software_number"))
	if err != nil {
		return domainFacility.BacnetObject{}, err
	}
	hardwareQuantity, err := parseUint8(row.get("hardware_quantity"))
	return domainFacility.BacnetObject{
		Base: base, FieldDeviceID: &ownerID, TextFix: row.get("text_fix"), Description: optionalString(row.get("description")),
		GMSVisible: parseBool(row.get("gms_visible")), Optional: parseBool(row.get("optional")), TextIndividual: optionalString(row.get("text_individual")),
		SoftwareType: domainFacility.BacnetSoftwareType(row.get("software_type")), SoftwareNumber: softwareNumber,
		HardwareType: domainFacility.BacnetHardwareType(row.get("hardware_type")), HardwareQuantity: hardwareQuantity,
		SoftwareReferenceID: optionalUUID(row.get("software_reference_id")), StateTextID: optionalUUID(row.get("state_text_id")),
		NotificationClassID: optionalUUID(row.get("notification_class_id")), AlarmTypeID: optionalUUID(row.get("alarm_type_id")),
		AlarmDefinitionID: optionalUUID(row.get("alarm_definition_id")),
	}, err
}

func scanSoftwareReferences(ctx context.Context, file *excelize.File, sink fielddeviceimport.Sink) error {
	batch := make([]fielddeviceimport.SoftwareReference, 0, importBatchSize)
	err := scanRows(file, "Data-SoftwareReferences", func(row rowValues) error {
		value, err := parseSoftwareReference(row)
		if err != nil {
			return err
		}
		batch = append(batch, value)
		if len(batch) == importBatchSize {
			err = sink.SoftwareReferences(ctx, batch)
			batch = make([]fielddeviceimport.SoftwareReference, 0, importBatchSize)
		}
		return err
	})
	if err != nil || len(batch) == 0 {
		return err
	}
	return sink.SoftwareReferences(ctx, batch)
}

func parseSoftwareReference(row rowValues) (fielddeviceimport.SoftwareReference, error) {
	sourceID, err := parseUUID(row.get("source_object_id"))
	if err != nil {
		return fielddeviceimport.SoftwareReference{}, err
	}
	targetID, err := parseUUID(row.get("target_object_id"))
	if err != nil {
		return fielddeviceimport.SoftwareReference{}, err
	}
	ownerID, err := parseUUID(row.get("field_device_id"))
	return fielddeviceimport.SoftwareReference{SourceObjectID: sourceID, TargetObjectID: targetID, FieldDeviceID: ownerID}, err
}

func scanAlarmValues(ctx context.Context, file *excelize.File, sink fielddeviceimport.Sink) error {
	batch := make([]domainFacility.BacnetObjectAlarmValue, 0, importBatchSize)
	err := scanRows(file, "Data-AlarmValues", func(row rowValues) error {
		value, err := parseAlarmValue(row)
		if err != nil {
			return err
		}
		batch = append(batch, value)
		if len(batch) == importBatchSize {
			err = sink.AlarmValues(ctx, batch)
			batch = make([]domainFacility.BacnetObjectAlarmValue, 0, importBatchSize)
		}
		return err
	})
	if err != nil || len(batch) == 0 {
		return err
	}
	return sink.AlarmValues(ctx, batch)
}

func parseAlarmValue(row rowValues) (domainFacility.BacnetObjectAlarmValue, error) {
	base, err := parseBase(row)
	if err != nil {
		return domainFacility.BacnetObjectAlarmValue{}, err
	}
	objectID, err := parseUUID(row.get("bacnet_object_id"))
	if err != nil {
		return domainFacility.BacnetObjectAlarmValue{}, err
	}
	fieldID, err := parseUUID(row.get("alarm_type_field_id"))
	return domainFacility.BacnetObjectAlarmValue{
		Base: base, BacnetObjectID: objectID, AlarmTypeFieldID: fieldID,
		ValueNumber: optionalFloat(row.get("value_number")), ValueInteger: optionalInt64(row.get("value_integer")),
		ValueBoolean: optionalBool(row.get("value_boolean")), ValueString: optionalString(row.get("value_string")),
		ValueJSON: optionalString(row.get("value_json")), UnitID: optionalUUID(row.get("unit_id")), Source: row.get("source"),
	}, err
}

func parseBase(row rowValues) (domain.Base, error) {
	id, err := parseUUID(row.get("source_id"))
	if err != nil {
		return domain.Base{}, err
	}
	version, err := strconv.ParseUint(row.get("version"), 10, 64)
	if err != nil {
		return domain.Base{}, err
	}
	createdAt, err := optionalTime(row.get("created_at"))
	if err != nil {
		return domain.Base{}, err
	}
	updatedAt, err := optionalTime(row.get("updated_at"))
	return domain.Base{ID: id, Version: version, CreatedAt: createdAt, UpdatedAt: updatedAt}, err
}

func optionalTime(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, value)
}

func parseUUID(value string) (uuid.UUID, error) { return uuid.Parse(strings.TrimSpace(value)) }

func optionalUUID(value string) *uuid.UUID {
	id, err := parseUUID(value)
	if err != nil {
		return nil
	}
	return &id
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalInt(value string) *int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func optionalInt64(value string) *int64 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func optionalFloat(value string) *float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func optionalBool(value string) *bool {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parsed := parseBool(value)
	return &parsed
}

func parseBool(value string) bool {
	parsed, _ := strconv.ParseBool(strings.TrimSpace(value))
	return parsed
}

func parseUint16(value string) (uint16, error) {
	parsed, err := strconv.ParseUint(value, 10, 16)
	return uint16(parsed), err
}

func parseUint8(value string) (uint8, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 8)
	return uint8(parsed), err
}
