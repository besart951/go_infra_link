package facilitybenchmark

import (
	"context"
	"time"
)

func (d *Database) seedFieldDevices(ctx context.Context) error {
	columns := baseColumns("bmk", "description", "apparat_nr", "text_individuell", "sps_controller_system_type_id", "system_part_id", "apparat_id")
	source := &generatedSource{count: d.FieldDeviceCount, row: fieldDeviceRow}
	return d.copy(ctx, "field_devices", columns, source)
}

func fieldDeviceRow(index int64) ([]any, error) {
	values := baseValues(8, index)
	var bmk any = "BMK-" + textValue(index)
	var description any = textValue(index)
	if index%5 == 1 {
		bmk = nil
	}
	if index%8 == 3 {
		description = nil
	}
	scope := index % SystemScopeCount
	values = append(values, bmk, description, int(index/SystemScopeCount)+1, nil,
		deterministicID(7, scope), deterministicID(3, scope%100), deterministicID(4, scope%100))
	return values, nil
}

func (d *Database) seedSpecifications(ctx context.Context) error {
	columns := baseColumns("field_device_id", "specification_supplier", "specification_brand", "specification_type", "additional_info_motor_valve", "additional_info_size", "additional_information_installation_location", "electrical_connection_ph", "electrical_connection_acdc", "electrical_connection_amperage", "electrical_connection_power", "electrical_connection_rotation")
	source := &generatedSource{count: d.FieldDeviceCount, row: specificationRow}
	return d.copy(ctx, "specifications", columns, source)
}

func (d *Database) seedCursorValues(ctx context.Context) error {
	_, err := d.PGX.Exec(ctx, benchmarkCursorValueBackfill)
	return err
}

const benchmarkCursorValueBackfill = `
INSERT INTO field_device_cursor_values (
  field_device_id,sps_number,sps_document_name,specification_supplier,specification_brand,
  specification_type,additional_info_motor_valve,additional_info_size,
  additional_information_installation_location,electrical_connection_ph,
  electrical_connection_acdc,electrical_connection_amperage,electrical_connection_power,
  electrical_connection_rotation
)
SELECT fd.id, scts.number, scts.document_name,
  specs.specification_supplier, specs.specification_brand, specs.specification_type,
  specs.additional_info_motor_valve, specs.additional_info_size,
  specs.additional_information_installation_location, specs.electrical_connection_ph,
  specs.electrical_connection_acdc, specs.electrical_connection_amperage,
  specs.electrical_connection_power, specs.electrical_connection_rotation
FROM field_devices fd
JOIN sps_controller_system_types scts ON scts.id=fd.sps_controller_system_type_id
LEFT JOIN specifications specs ON specs.field_device_id=fd.id
ON CONFLICT (field_device_id) DO UPDATE SET
  sps_number=EXCLUDED.sps_number, sps_document_name=EXCLUDED.sps_document_name,
  specification_supplier=EXCLUDED.specification_supplier,
  specification_brand=EXCLUDED.specification_brand, specification_type=EXCLUDED.specification_type,
  additional_info_motor_valve=EXCLUDED.additional_info_motor_valve,
  additional_info_size=EXCLUDED.additional_info_size,
  additional_information_installation_location=EXCLUDED.additional_information_installation_location,
  electrical_connection_ph=EXCLUDED.electrical_connection_ph,
  electrical_connection_acdc=EXCLUDED.electrical_connection_acdc,
  electrical_connection_amperage=EXCLUDED.electrical_connection_amperage,
  electrical_connection_power=EXCLUDED.electrical_connection_power,
  electrical_connection_rotation=EXCLUDED.electrical_connection_rotation`

func specificationRow(index int64) ([]any, error) {
	var supplier any = "Supplier " + textValue(index)
	var brand any = "Brand"
	if index%6 == 0 {
		supplier = nil
	}
	if index%9 == 0 {
		brand = nil
	}
	values := baseValues(9, index)
	values = append(values, deterministicID(8, index), supplier, brand, "Type", nil, int(index%200), nil, nil, nil, nil, nil, nil)
	return values, nil
}

func (d *Database) seedProjectLinks(ctx context.Context) error {
	columns := baseColumns("project_id", "field_device_id", "field_device_created_at")
	source := &generatedSource{count: d.FieldDeviceCount, row: func(index int64) ([]any, error) {
		createdAt := benchmarkEpoch.Add(time.Duration(index) * time.Microsecond)
		return append(baseValues(10, index), deterministicID(11, index%ProjectCount), deterministicID(8, index), createdAt), nil
	}}
	return d.copy(ctx, "project_field_devices", columns, source)
}
