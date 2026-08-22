package exporting

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

var softwareKeys = []string{"ai", "ao", "av", "bi", "bo", "bv", "mi", "mo", "mv", "ca", "ee", "lp", "nc", "sc", "tl"}
var hardwareKeys = []string{"do", "ao", "di", "ai"}

var headings = []string{
	"BACnet Object Name",
	"Description",
	"State Text (Tabelle der Zustandstexte) / (B2) /",
	"Notification",
	"BMK",
	"GMS Sichtbar",
	"Anlageteil",
	"Apparat",
	"Anlageteil (Abkürzung)",
	"Apparat (Abkürzung)",
	"Text-Fix",
	"AI", "AO", "AV", "BI", "BO", "BV", "MI", "MO", "MV", "CA", "EE", "LP", "NC", "SC", "TL",
	"Adresse",
	"Alarm/Definition",
	"State Text",
	"DO", "AO", "DI", "AI",
	"Bemerkung",
	"Lieferant",
	"Fabrikat",
	"Typ",
	"Motor,Ventil, etc.",
	"Grösse",
	"Montageort",
	"Ph",
	"AC/DC",
	"Stromstärke",
	"Leistung",
	"Drehzahl",
}

// styles holds pre-created excelize style IDs for the workbook.
type styles struct {
	headerTitle    int // Row 1: bold, size 16
	headerInfo     int // Rows 2-10: bold
	columnHeading  int // Heading row: bold, size 12, light gray background
	firstLineStyle int // Device first-line: bold, light blue background
}

func createStyles(f *excelize.File) (styles, error) {
	var s styles
	var err error

	s.headerTitle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
	})
	if err != nil {
		return s, err
	}

	s.headerInfo, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return s, err
	}

	s.columnHeading, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F0F0F0"}},
	})
	if err != nil {
		return s, err
	}

	s.firstLineStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"ADD8E6"}},
	})
	if err != nil {
		return s, err
	}

	return s, nil
}

type ExcelizeGenerator struct{}

func NewExcelizeGenerator() *ExcelizeGenerator {
	return &ExcelizeGenerator{}
}

const excelMaximumRows = 1_048_576

func (g *ExcelizeGenerator) GenerateWorkbook(ctx context.Context, outputPath string, controllers []domainExport.Controller, source domainExport.DataProvider, req domainExport.Request, pageSize int) (int64, error) {
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 500
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	st, err := createStyles(f)
	if err != nil {
		return 0, err
	}

	defaultSheet := f.GetSheetName(0)
	var written int64
	usedSheets := make(map[string]struct{})
	if err := writeExportManifest(f, req); err != nil {
		return 0, err
	}
	machineSheets, err := newMachineDataSheets(f)
	if err != nil {
		return 0, err
	}

	for idx, controller := range sortedControllers(controllers) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		part := 1
		sheetName := uniqueSheetName(safeSheetName(controller.GADevice, controller.ID), part, usedSheets)
		if idx == 0 && defaultSheet != "" {
			if err := f.SetSheetName(defaultSheet, sheetName); err != nil {
				return 0, err
			}
		} else {
			if _, err = f.NewSheet(sheetName); err != nil {
				return 0, err
			}
		}
		stream, err := f.NewStreamWriter(sheetName)
		if err != nil {
			return 0, err
		}
		rowIdx, err := writeControllerHeading(stream, controller, st)
		if err != nil {
			return 0, err
		}

		afterID := uuid.Nil
		for {
			devices, listErr := source.ListFieldDevicesByControllerAfter(ctx, controller.ID, req, afterID, pageSize)
			if listErr != nil {
				return 0, listErr
			}
			if len(devices) == 0 {
				break
			}
			for _, device := range devices {
				rowsNeeded := 1 + len(device.BacnetObjects)
				if rowIdx+rowsNeeded-1 > excelMaximumRows {
					if err := stream.Flush(); err != nil {
						return 0, err
					}
					part++
					sheetName = uniqueSheetName(safeSheetName(controller.GADevice, controller.ID), part, usedSheets)
					if _, err := f.NewSheet(sheetName); err != nil {
						return 0, err
					}
					stream, err = f.NewStreamWriter(sheetName)
					if err != nil {
						return 0, err
					}
					rowIdx, err = writeControllerHeading(stream, controller, st)
					if err != nil {
						return 0, err
					}
				}
				if err := stream.SetRow(cell("A", rowIdx), styledAnyRow(firstLine(controller, device), st.firstLineStyle)); err != nil {
					return 0, err
				}
				rowIdx++

				for _, bo := range device.BacnetObjects {
					if err := stream.SetRow(cell("A", rowIdx), anyToCells(bacnetLine(controller, device, bo))); err != nil {
						return 0, err
					}
					rowIdx++
				}
				written++
				if err := machineSheets.Write(controller, device); err != nil {
					return 0, err
				}
			}
			afterID = devices[len(devices)-1].ID
			if len(devices) < pageSize {
				break
			}
		}

		if err := stream.Flush(); err != nil {
			return 0, err
		}
	}
	if err := machineSheets.Close(); err != nil {
		return 0, err
	}

	if len(f.GetSheetList()) > 0 {
		f.SetActiveSheet(0)
	}

	return written, f.SaveAs(outputPath)
}

type dataSheetWriter struct {
	file     *excelize.File
	baseName string
	headings []string
	part     int
	row      int
	stream   *excelize.StreamWriter
}

func newDataSheetWriter(file *excelize.File, baseName string, headings []string) (*dataSheetWriter, error) {
	w := &dataSheetWriter{file: file, baseName: baseName, headings: headings}
	if err := w.nextSheet(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *dataSheetWriter) Write(values []any) error {
	if w.row > excelMaximumRows {
		if err := w.stream.Flush(); err != nil {
			return err
		}
		if err := w.nextSheet(); err != nil {
			return err
		}
	}
	if err := w.stream.SetRow(cell("A", w.row), anyToCells(values)); err != nil {
		return err
	}
	w.row++
	return nil
}

func (w *dataSheetWriter) nextSheet() error {
	w.part++
	name := w.baseName
	if w.part > 1 {
		name = fmt.Sprintf("%s-%d", w.baseName, w.part)
	}
	if _, err := w.file.NewSheet(name); err != nil {
		return err
	}
	stream, err := w.file.NewStreamWriter(name)
	if err != nil {
		return err
	}
	w.stream = stream
	w.row = 2
	return stream.SetRow("A1", anyToCells(stringSliceToAny(w.headings)))
}

func (w *dataSheetWriter) Close() error {
	return w.stream.Flush()
}

type machineDataSheets struct {
	fieldDevices   *dataSheetWriter
	specifications *dataSheetWriter
	bacnetObjects  *dataSheetWriter
	softwareRefs   *dataSheetWriter
	alarmValues    *dataSheetWriter
}

func newMachineDataSheets(file *excelize.File) (*machineDataSheets, error) {
	fieldDevices, err := newDataSheetWriter(file, "Data-FieldDevices", []string{
		"source_id", "version", "created_at", "updated_at", "sps_controller_id", "sps_controller_system_type_id",
		"system_part_id", "apparat_id", "apparat_nr", "bmk", "description", "text_individual",
	})
	if err != nil {
		return nil, err
	}
	specifications, err := newDataSheetWriter(file, "Data-Specifications", []string{
		"source_id", "field_device_id", "version", "supplier", "brand", "type", "motor_valve", "size",
		"installation_location", "ph", "acdc", "amperage", "power", "rotation",
	})
	if err != nil {
		return nil, err
	}
	bacnetObjects, err := newDataSheetWriter(file, "Data-BACnetObjects", []string{
		"source_id", "field_device_id", "version", "text_fix", "description", "gms_visible", "optional", "text_individual",
		"software_type", "software_number", "hardware_type", "hardware_quantity", "software_reference_id",
		"state_text_id", "notification_class_id", "alarm_type_id", "alarm_definition_id",
	})
	if err != nil {
		return nil, err
	}
	softwareRefs, err := newDataSheetWriter(file, "Data-SoftwareReferences", []string{
		"source_object_id", "target_object_id", "field_device_id",
	})
	if err != nil {
		return nil, err
	}
	alarmValues, err := newDataSheetWriter(file, "Data-AlarmValues", []string{
		"source_id", "bacnet_object_id", "version", "alarm_type_field_id", "value_number", "value_integer",
		"value_boolean", "value_string", "value_json", "unit_id", "source",
	})
	if err != nil {
		return nil, err
	}
	return &machineDataSheets{
		fieldDevices: fieldDevices, specifications: specifications, bacnetObjects: bacnetObjects,
		softwareRefs: softwareRefs, alarmValues: alarmValues,
	}, nil
}

func (s *machineDataSheets) Write(controller domainExport.Controller, device domainFacility.FieldDevice) error {
	if err := s.fieldDevices.Write([]any{
		device.ID.String(), device.Version, device.CreatedAt, device.UpdatedAt, controller.ID.String(), device.SPSControllerSystemTypeID.String(),
		device.SystemPartID.String(), device.ApparatID.String(), device.ApparatNr, strPtr(device.BMK), strPtr(device.Description), strPtr(device.TextIndividuell),
	}); err != nil {
		return err
	}
	if spec := device.Specification; spec != nil {
		if err := s.specifications.Write([]any{
			spec.ID.String(), device.ID.String(), spec.Version, strPtr(spec.SpecificationSupplier), strPtr(spec.SpecificationBrand),
			strPtr(spec.SpecificationType), strPtr(spec.AdditionalInfoMotorValve), intPtr(spec.AdditionalInfoSize),
			strPtr(spec.AdditionalInformationInstallationLocation), intPtr(spec.ElectricalConnectionPH), strPtr(spec.ElectricalConnectionACDC),
			floatPtr(spec.ElectricalConnectionAmperage), floatPtr(spec.ElectricalConnectionPower), intPtr(spec.ElectricalConnectionRotation),
		}); err != nil {
			return err
		}
	}
	for _, object := range device.BacnetObjects {
		if err := s.bacnetObjects.Write([]any{
			object.ID.String(), device.ID.String(), object.Version, object.TextFix, strPtr(object.Description), object.GMSVisible, object.Optional,
			strPtr(object.TextIndividual), object.SoftwareType, object.SoftwareNumber, object.HardwareType, object.HardwareQuantity,
			uuidPtr(object.SoftwareReferenceID), uuidPtr(object.StateTextID), uuidPtr(object.NotificationClassID), uuidPtr(object.AlarmTypeID), uuidPtr(object.AlarmDefinitionID),
		}); err != nil {
			return err
		}
		if object.SoftwareReferenceID != nil {
			if err := s.softwareRefs.Write([]any{
				object.ID.String(), object.SoftwareReferenceID.String(), device.ID.String(),
			}); err != nil {
				return err
			}
		}
		for _, value := range object.AlarmValues {
			if err := s.alarmValues.Write([]any{
				value.ID.String(), object.ID.String(), value.Version, value.AlarmTypeFieldID.String(), pointerValue(value.ValueNumber),
				pointerValue(value.ValueInteger), pointerValue(value.ValueBoolean), pointerValue(value.ValueString), pointerValue(value.ValueJSON),
				uuidPtr(value.UnitID), value.Source,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *machineDataSheets) Close() error {
	return errors.Join(
		s.fieldDevices.Close(), s.specifications.Close(), s.bacnetObjects.Close(),
		s.softwareRefs.Close(), s.alarmValues.Close(),
	)
}

func writeExportManifest(file *excelize.File, req domainExport.Request) error {
	if _, err := file.NewSheet("Export-Manifest"); err != nil {
		return err
	}
	stream, err := file.NewStreamWriter("Export-Manifest")
	if err != nil {
		return err
	}
	for index, row := range exportManifestRows(req) {
		if err := stream.SetRow(cell("A", index+1), anyToCells(row)); err != nil {
			return err
		}
	}
	return stream.Flush()
}

func exportManifestRows(req domainExport.Request) [][]any {
	checksums, _ := json.Marshal(req.Manifest.SnapshotChecksums)
	return [][]any{
		{"schema_version", req.SchemaVersion},
		{"snapshot_at", req.SnapshotAt.UTC().Format(time.RFC3339Nano)},
		{"scope", req.AccessScope},
		{"device_count", req.DeviceCount},
		{"specification_count", req.Manifest.Counts.Specifications},
		{"bacnet_object_count", req.Manifest.Counts.BacnetObjects},
		{"software_reference_count", req.Manifest.Counts.SoftwareReferences},
		{"alarm_value_count", req.Manifest.Counts.AlarmValues},
		{"project_ids", joinUUIDs(req.ProjectIDs)},
		{"building_ids", joinUUIDs(req.BuildingIDs)},
		{"control_cabinet_ids", joinUUIDs(req.ControlCabinetIDs)},
		{"sps_controller_ids", joinUUIDs(req.SPSControllerIDs)},
		{"sps_controller_system_type_ids", joinUUIDs(req.SPSControllerSystemTypeIDs)},
		{"search", req.Search},
		{"workbook_shards", strings.Join(req.Manifest.WorkbookShards, ",")},
		{"warnings", strings.Join(req.Manifest.Warnings, " | ")},
		{"snapshot_checksums", string(checksums)},
	}
}

func stringSliceToAny(values []string) []any {
	out := make([]any, len(values))
	for i := range values {
		out[i] = values[i]
	}
	return out
}

func joinUUIDs(values []uuid.UUID) string {
	parts := make([]string, len(values))
	for i := range values {
		parts[i] = values[i].String()
	}
	return strings.Join(parts, ",")
}

func uuidPtr(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func intPtr(value *int) any {
	if value == nil {
		return ""
	}
	return *value
}

func floatPtr(value *float64) any {
	if value == nil {
		return ""
	}
	return *value
}

func pointerValue[T any](value *T) any {
	if value == nil {
		return ""
	}
	return *value
}

func writeControllerHeading(stream *excelize.StreamWriter, controller domainExport.Controller, st styles) (int, error) {
	rowIdx := 1
	for i, row := range controllerHeaderRows(controller) {
		styleID := st.headerInfo
		if i == 0 {
			styleID = st.headerTitle
		}
		if err := stream.SetRow(cell("A", rowIdx), styledRow(row, styleID)); err != nil {
			return 0, err
		}
		rowIdx++
	}
	if err := stream.SetRow(cell("A", rowIdx), []any{excelize.Cell{Value: " "}}); err != nil {
		return 0, err
	}
	rowIdx++
	if err := stream.SetRow(cell("A", rowIdx), styledRow(headings, st.columnHeading)); err != nil {
		return 0, err
	}
	return rowIdx + 1, nil
}

func uniqueSheetName(base string, part int, used map[string]struct{}) string {
	candidate := base
	if part > 1 {
		suffix := fmt.Sprintf("-%d", part)
		maxBase := max(1, 31-len(suffix))
		if len(candidate) > maxBase {
			candidate = candidate[:maxBase]
		}
		candidate += suffix
	}
	for n := 2; ; n++ {
		if _, exists := used[candidate]; !exists {
			used[candidate] = struct{}{}
			return candidate
		}
		suffix := fmt.Sprintf("-%d", n)
		trimmed := base
		if len(trimmed) > 31-len(suffix) {
			trimmed = trimmed[:31-len(suffix)]
		}
		candidate = trimmed + suffix
	}
}

func (g *ExcelizeGenerator) GenerateZipByCabinet(ctx context.Context, outputPath string, controllers []domainExport.Controller, source domainExport.DataProvider, req domainExport.Request, pageSize int) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	defer func() { _ = zw.Close() }()

	byCabinet := map[uuid.UUID][]domainExport.Controller{}
	for _, controller := range controllers {
		byCabinet[controller.ControlCabinetID] = append(byCabinet[controller.ControlCabinetID], controller)
	}

	usedEntryNames := map[string]struct{}{}
	var written int64

	for _, cabinetID := range sortedCabinetIDs(byCabinet) {
		cabinetControllers := byCabinet[cabinetID]
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		tmp, err := os.CreateTemp("", "field-device-export-*.xlsx")
		if err != nil {
			return 0, err
		}
		tmpPath := tmp.Name()
		_ = tmp.Close()

		count, err := g.GenerateWorkbook(ctx, tmpPath, cabinetControllers, source, req, pageSize)
		if err != nil {
			_ = os.Remove(tmpPath)
			return 0, err
		}
		written += count

		entryName := safeCabinetFileName(cabinetControllers[0], cabinetID)
		entryName = ensureUniqueZipEntryName(entryName, usedEntryNames)
		entry, err := zw.Create(entryName)
		if err != nil {
			_ = os.Remove(tmpPath)
			return 0, err
		}

		content, err := os.Open(tmpPath)
		if err != nil {
			_ = os.Remove(tmpPath)
			return 0, err
		}
		_, copyErr := io.Copy(entry, content)
		closeErr := content.Close()
		if copyErr != nil {
			_ = os.Remove(tmpPath)
			return 0, copyErr
		}
		if closeErr != nil {
			_ = os.Remove(tmpPath)
			return 0, closeErr
		}

		_ = os.Remove(tmpPath)
	}

	return written, zw.Close()
}

func sortedCabinetIDs(items map[uuid.UUID][]domainExport.Controller) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left].String() < ids[right].String() })
	return ids
}

// ---------------------------------------------------------------------------
// Controller header
// ---------------------------------------------------------------------------

func controllerHeaderRows(ctrl domainExport.Controller) [][]string {
	bgStr := fmt.Sprintf("%d", ctrl.BuildingGroup)
	schaltschrankNr := strings.Join(filterEmpty([]string{bgStr, ctrl.MinSystemPartNumber, "00"}), "_")

	return [][]string{
		{"Projekt Controller", ctrl.GADevice},
		{"GA-Gerät:", ctrl.GADevice},
		{"Schaltschrank-Nr.", schaltschrankNr},
		{"Device Name:", ctrl.DeviceName},
		{"Device Instance:", ctrl.DeviceInstance},
		{"Device Description:", ctrl.DeviceDescription},
		{"Device Location:", ctrl.DeviceLocation},
		{"IP-Adresse:", ctrl.IPAddress},
		{"Subnetz:", ctrl.Subnet},
		{"Gateway:", ctrl.Gateway},
		{"VLAN:", ctrl.VLAN},
	}
}

// ---------------------------------------------------------------------------
// Data row builders
// ---------------------------------------------------------------------------

func firstLine(ctrl domainExport.Controller, device domainFacility.FieldDevice) []any {
	softwareSums := map[string]float64{}
	hardwareSums := map[string]float64{}
	for _, key := range softwareKeys {
		softwareSums[key] = 0
	}
	for _, key := range hardwareKeys {
		hardwareSums[key] = 0
	}

	for _, bo := range device.BacnetObjects {
		s := softwareMetrics(bo)
		h := hardwareMetrics(bo)
		for _, key := range softwareKeys {
			softwareSums[key] += s[key]
		}
		for _, key := range hardwareKeys {
			hardwareSums[key] += h[key]
		}
	}

	gmsVisible := false
	if len(device.BacnetObjects) > 0 {
		gmsVisible = device.BacnetObjects[0].GMSVisible
	}

	row := []any{
		buildBacnetObjectName(ctrl, device, ""),
		buildFieldDeviceDescription(device),
		"",
		"",
		strPtr(device.BMK),
		gmsVisible,
		device.SystemPart.Name,
		device.Apparat.Name,
		device.SystemPart.ShortName,
		device.Apparat.ShortName,
		"",
	}

	for _, key := range softwareKeys {
		row = append(row, softwareSums[key])
	}

	row = append(row, "", "", "")

	for _, key := range hardwareKeys {
		row = append(row, hardwareSums[key])
	}

	row = append(row,
		strPtr(device.Description),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.SpecificationSupplier }),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.SpecificationBrand }),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.SpecificationType }),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.AdditionalInfoMotorValve }),
		specInt(device.Specification, func(s *domainFacility.Specification) *int { return s.AdditionalInfoSize }),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.AdditionalInformationInstallationLocation }),
		specInt(device.Specification, func(s *domainFacility.Specification) *int { return s.ElectricalConnectionPH }),
		specString(device.Specification, func(s *domainFacility.Specification) *string { return s.ElectricalConnectionACDC }),
		specFloat(device.Specification, func(s *domainFacility.Specification) *float64 { return s.ElectricalConnectionAmperage }),
		specFloat(device.Specification, func(s *domainFacility.Specification) *float64 { return s.ElectricalConnectionPower }),
		specInt(device.Specification, func(s *domainFacility.Specification) *int { return s.ElectricalConnectionRotation }),
	)

	return row
}

func bacnetLine(ctrl domainExport.Controller, device domainFacility.FieldDevice, bo domainFacility.BacnetObject) []any {
	s := softwareMetrics(bo)
	h := hardwareMetrics(bo)
	address := softwareAddress(bo)

	row := []any{
		buildBacnetObjectName(ctrl, device, address),
		buildDescription(device, bo.TextFix),
		aggregateStateTexts(bo.StateText),
		notificationNC(bo.NotificationClass),
		strPtr(device.BMK),
		bo.GMSVisible,
		device.SystemPart.Name,
		device.Apparat.Name,
		device.SystemPart.ShortName,
		device.Apparat.ShortName,
		bo.TextFix,
	}

	for _, key := range softwareKeys {
		row = append(row, s[key])
	}

	row = append(row,
		address,
		alarmName(bo.AlarmType),
		firstStateText(bo.StateText),
	)

	for _, key := range hardwareKeys {
		row = append(row, h[key])
	}

	row = append(row, "", "", "", "", "", "", "", "", "", "", "", "")
	return row
}

// ---------------------------------------------------------------------------
// BACnet helpers
// ---------------------------------------------------------------------------

func softwareMetrics(bo domainFacility.BacnetObject) map[string]float64 {
	out := map[string]float64{}
	for _, key := range softwareKeys {
		out[key] = 0
	}
	key := strings.ToLower(string(bo.SoftwareType))
	if _, ok := out[key]; ok {
		out[key] = 1
	}
	return out
}

func softwareAddress(bo domainFacility.BacnetObject) string {
	key := strings.ToUpper(string(bo.SoftwareType))
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s%02d", key, bo.SoftwareNumber)
}

func hardwareMetrics(bo domainFacility.BacnetObject) map[string]float64 {
	out := map[string]float64{}
	for _, key := range hardwareKeys {
		out[key] = 0
	}
	key := strings.ToLower(string(bo.HardwareType))
	if _, ok := out[key]; ok {
		out[key] = float64(bo.HardwareQuantity)
	}
	return out
}

func buildBacnetObjectName(ctrl domainExport.Controller, device domainFacility.FieldDevice, suffix string) string {
	sysTypeNr := ""
	if device.SPSControllerSystemType.Number != nil {
		sysTypeNr = fmt.Sprintf("%04d", *device.SPSControllerSystemType.Number)
	}

	devicePart := ""
	if device.SystemPart.ShortName != "" || device.Apparat.ShortName != "" {
		devicePart = device.SystemPart.ShortName + device.Apparat.ShortName + fmt.Sprintf("%02d", device.ApparatNr)
	}

	nameParts := filterEmpty([]string{
		ctrl.IWSCode,
		fmt.Sprintf("%d", ctrl.BuildingGroup),
		sysTypeNr,
		ctrl.GADevice,
		devicePart,
	})

	base := strings.Join(nameParts, "_")
	if suffix == "" {
		return base
	}
	if base == "" {
		return suffix
	}
	return base + "_" + suffix
}

func buildDescription(device domainFacility.FieldDevice, textFix string) string {
	parts := []string{strings.TrimSpace(device.SystemPart.Name), strings.TrimSpace(device.Apparat.Name)}
	left := strings.TrimSpace(strings.Join(parts, " "))
	short := strings.TrimSpace(device.SystemPart.ShortName + device.Apparat.ShortName)
	msg := strings.TrimSpace(textFix)
	out := strings.TrimSpace(left + " - " + short)
	if msg != "" {
		out = strings.TrimSpace(out + " " + msg)
	}
	return strings.TrimSpace(out)
}

func aggregateStateTexts(st *domainFacility.StateText) string {
	if st == nil {
		return ""
	}
	items := []string{}
	vals := []*string{st.StateText1, st.StateText2, st.StateText3, st.StateText4, st.StateText5, st.StateText6, st.StateText7, st.StateText8, st.StateText9, st.StateText10, st.StateText11, st.StateText12, st.StateText13, st.StateText14, st.StateText15, st.StateText16}
	for _, v := range vals {
		if v != nil && strings.TrimSpace(*v) != "" {
			items = append(items, strings.TrimSpace(*v))
		}
	}
	return strings.Join(items, ", ")
}

func firstStateText(st *domainFacility.StateText) string {
	if st == nil || st.StateText1 == nil {
		return ""
	}
	return *st.StateText1
}

func notificationNC(nc *domainFacility.NotificationClass) any {
	if nc == nil {
		return ""
	}
	return nc.Nc
}

func alarmName(alarmType *domainFacility.AlarmType) string {
	if alarmType == nil {
		return ""
	}
	return alarmType.Name
}

// ---------------------------------------------------------------------------
// Specification helpers
// ---------------------------------------------------------------------------

func strPtr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func specString(spec *domainFacility.Specification, getter func(*domainFacility.Specification) *string) string {
	if spec == nil {
		return ""
	}
	v := getter(spec)
	if v == nil {
		return ""
	}
	return *v
}

func specInt(spec *domainFacility.Specification, getter func(*domainFacility.Specification) *int) any {
	if spec == nil {
		return ""
	}
	v := getter(spec)
	if v == nil {
		return ""
	}
	return *v
}

func specFloat(spec *domainFacility.Specification, getter func(*domainFacility.Specification) *float64) any {
	if spec == nil {
		return ""
	}
	v := getter(spec)
	if v == nil {
		return ""
	}
	return *v
}

// ---------------------------------------------------------------------------
// Sorting / naming
// ---------------------------------------------------------------------------

func sortedControllers(controllers []domainExport.Controller) []domainExport.Controller {
	out := append([]domainExport.Controller{}, controllers...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ControlCabinetID == out[j].ControlCabinetID {
			return out[i].GADevice < out[j].GADevice
		}
		return out[i].ControlCabinetID.String() < out[j].ControlCabinetID.String()
	})
	return out
}

func buildFieldDeviceDescription(device domainFacility.FieldDevice) string {
	parts := filterEmpty([]string{
		strings.TrimSpace(device.SystemPart.Name),
		strings.TrimSpace(device.Apparat.Name),
		strings.TrimSpace(strPtr(device.TextIndividuell)),
	})
	return strings.Join(parts, " ")
}

func safeCabinetFileName(ctrl domainExport.Controller, cabinetID uuid.UUID) string {
	name := strings.Join(filterEmpty([]string{
		strings.TrimSpace(ctrl.IWSCode),
		fmt.Sprintf("%d", ctrl.BuildingGroup),
		strings.TrimSpace(ctrl.ControlCabinetNr),
	}), "_")

	if name == "" {
		name = "cabinet-" + ctrl.ControlCabinetID.String()[:8]
	}

	invalid := []string{"\\", "/", "*", "?", ":", "[", "]", "<", ">", "|", "\""}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "-")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "cabinet-" + cabinetID.String()[:8]
	}

	if len(name) > 120 {
		name = name[:120]
	}

	return name + ".xlsx"
}

func ensureUniqueZipEntryName(base string, used map[string]struct{}) string {
	if _, exists := used[base]; !exists {
		used[base] = struct{}{}
		return base
	}

	for i := 2; ; i++ {
		ext := filepath.Ext(base)
		nameOnly := strings.TrimSuffix(base, ext)
		candidate := fmt.Sprintf("%s-%d%s", nameOnly, i, ext)
		if _, exists := used[candidate]; !exists {
			used[candidate] = struct{}{}
			return candidate
		}
	}
}

func safeSheetName(ga string, id uuid.UUID) string {
	name := "Projekt Controller " + strings.TrimSpace(ga)
	if strings.TrimSpace(ga) == "" {
		name = "controller-" + id.String()[:8]
	}
	invalid := []string{"\\", "/", "*", "?", ":", "[", "]"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "-")
	}
	if len(name) > 31 {
		name = name[:31]
	}
	return name
}

// ---------------------------------------------------------------------------
// Cell / row helpers for excelize StreamWriter
// ---------------------------------------------------------------------------

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}

func styledRow(values []string, styleID int) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, excelize.Cell{StyleID: styleID, Value: v})
	}
	return out
}

func styledAnyRow(values []any, styleID int) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, excelize.Cell{StyleID: styleID, Value: v})
	}
	return out
}

func anyToCells(values []any) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, excelize.Cell{Value: v})
	}
	return out
}
