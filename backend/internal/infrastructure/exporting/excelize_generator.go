package exporting

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func (g *ExcelizeGenerator) GenerateWorkbook(ctx context.Context, outputPath string, controllers []domainExport.Controller, perControllerDevices map[uuid.UUID][]domainFacility.FieldDevice) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	st, err := createStyles(f)
	if err != nil {
		return err
	}

	defaultSheet := f.GetSheetName(0)
	if defaultSheet != "" {
		f.DeleteSheet(defaultSheet)
	}

	for _, controller := range sortedControllers(controllers) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sheetName := safeSheetName(controller.GADevice, controller.ID)
		_, _ = f.NewSheet(sheetName)
		stream, err := f.NewStreamWriter(sheetName)
		if err != nil {
			return err
		}

		rowIdx := 1

		// Controller header rows
		headerRows := controllerHeaderRows(controller)
		for i, row := range headerRows {
			styleID := st.headerInfo
			if i == 0 {
				styleID = st.headerTitle
			}
			if err := stream.SetRow(cell("A", rowIdx), styledRow(row, styleID)); err != nil {
				return err
			}
			rowIdx++
		}

		// Blank separator row
		if err := stream.SetRow(cell("A", rowIdx), []any{excelize.Cell{Value: " "}}); err != nil {
			return err
		}
		rowIdx++

		// Column headings
		if err := stream.SetRow(cell("A", rowIdx), styledRow(headings, st.columnHeading)); err != nil {
			return err
		}
		rowIdx++

		// Data rows
		for _, device := range perControllerDevices[controller.ID] {
			if err := stream.SetRow(cell("A", rowIdx), styledAnyRow(firstLine(controller, device), st.firstLineStyle)); err != nil {
				return err
			}
			rowIdx++

			for _, bo := range device.BacnetObjects {
				if err := stream.SetRow(cell("A", rowIdx), anyToCells(bacnetLine(controller, device, bo))); err != nil {
					return err
				}
				rowIdx++
			}
		}

		if err := stream.Flush(); err != nil {
			return err
		}
	}

	if len(f.GetSheetList()) > 0 {
		f.SetActiveSheet(0)
	}

	return f.SaveAs(outputPath)
}

func (g *ExcelizeGenerator) GenerateZipByCabinet(ctx context.Context, outputPath string, controllers []domainExport.Controller, perControllerDevices map[uuid.UUID][]domainFacility.FieldDevice) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	defer func() { _ = zw.Close() }()

	byCabinet := map[uuid.UUID][]domainExport.Controller{}
	for _, controller := range controllers {
		byCabinet[controller.ControlCabinetID] = append(byCabinet[controller.ControlCabinetID], controller)
	}

	for cabinetID, cabinetControllers := range byCabinet {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tmp, err := os.CreateTemp("", "field-device-export-*.xlsx")
		if err != nil {
			return err
		}
		tmpPath := tmp.Name()
		_ = tmp.Close()

		if err := g.GenerateWorkbook(ctx, tmpPath, cabinetControllers, perControllerDevices); err != nil {
			_ = os.Remove(tmpPath)
			return err
		}

		entryName := fmt.Sprintf("control-cabinet-%s.xlsx", cabinetID.String())
		entry, err := zw.Create(entryName)
		if err != nil {
			_ = os.Remove(tmpPath)
			return err
		}

		content, err := os.ReadFile(tmpPath)
		if err != nil {
			_ = os.Remove(tmpPath)
			return err
		}
		if _, err := entry.Write(content); err != nil {
			_ = os.Remove(tmpPath)
			return err
		}

		_ = os.Remove(tmpPath)
	}

	return zw.Close()
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
		buildDescription(device, ""),
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
		alarmName(bo.AlarmDefinition),
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

func alarmName(ad *domainFacility.AlarmDefinition) string {
	if ad == nil {
		return ""
	}
	return ad.Name
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
