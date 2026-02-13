package exporting

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

type ExcelizeGenerator struct{}

func NewExcelizeGenerator() *ExcelizeGenerator {
	return &ExcelizeGenerator{}
}

func (g *ExcelizeGenerator) GenerateWorkbook(ctx context.Context, outputPath string, controllers []domainExport.Controller, perControllerDevices map[uuid.UUID][]domainFacility.FieldDevice) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

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
		for _, row := range controllerHeaderRows(controller) {
			if err := stream.SetRow(fmt.Sprintf("A%d", rowIdx), toRow(row)); err != nil {
				return err
			}
			rowIdx++
		}
		if err := stream.SetRow(fmt.Sprintf("A%d", rowIdx), toRow([]string{" "})); err != nil {
			return err
		}
		rowIdx++

		if err := stream.SetRow(fmt.Sprintf("A%d", rowIdx), toRow(headings)); err != nil {
			return err
		}
		rowIdx++

		for _, device := range perControllerDevices[controller.ID] {
			if err := stream.SetRow(fmt.Sprintf("A%d", rowIdx), toAnyRow(firstLine(device))); err != nil {
				return err
			}
			rowIdx++

			for _, bo := range device.BacnetObjects {
				if err := stream.SetRow(fmt.Sprintf("A%d", rowIdx), toAnyRow(bacnetLine(device, bo))); err != nil {
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

func controllerHeaderRows(controller domainExport.Controller) [][]string {
	return [][]string{
		{"Projekt Controller", controller.GADevice},
		{"GA-Gerät:", controller.GADevice},
		{"Schaltschrank-Nr.", controller.ControlCabinetID.String()},
		{"Device Name:", ""},
		{"Device Instance:", ""},
		{"Device Description:", ""},
		{"Device Location:", ""},
		{"IP-Adresse:", ""},
		{"Subnetz:", ""},
		{"Gateway:", ""},
		{"VLAN:", ""},
	}
}

func firstLine(device domainFacility.FieldDevice) []any {
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
		buildBacnetObjectName(device, ""),
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

func bacnetLine(device domainFacility.FieldDevice, bo domainFacility.BacnetObject) []any {
	s := softwareMetrics(bo)
	h := hardwareMetrics(bo)
	address := softwareAddress(bo)

	row := []any{
		buildBacnetObjectName(device, address),
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
	return key + strconv.FormatInt(int64(bo.SoftwareNumber), 10)
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

func buildBacnetObjectName(device domainFacility.FieldDevice, suffix string) string {
	parts := []string{}
	if sysController := device.SPSControllerSystemType.SPSController; sysController.GADevice != nil {
		parts = append(parts, *sysController.GADevice)
	}
	if device.SystemPart.ShortName != "" || device.Apparat.ShortName != "" {
		parts = append(parts, device.SystemPart.ShortName+device.Apparat.ShortName+fmt.Sprintf("%02d", device.ApparatNr))
	}
	base := strings.Join(parts, "_")
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
	name := strings.TrimSpace(ga)
	if name == "" {
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

func toRow(values []string) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, v)
	}
	return out
}

func toAnyRow(values []any) []any {
	return values
}
