import type { FieldDevice, BacnetObject, Specification } from '$lib/domain/facility/index.js';

export function buildSpecification(overrides: Partial<Specification> = {}): Specification {
	return {
		id: 'spec-1',
		field_device_id: 'fd-1',
		specification_supplier: 'Supplier A',
		specification_brand: 'Brand A',
		specification_type: 'Type A',
		additional_info_motor_valve: 'Valve X',
		additional_info_size: 20,
		additional_information_installation_location: 'Roof',
		electrical_connection_ph: 3,
		electrical_connection_acdc: 'AC',
		electrical_connection_amperage: 2.5,
		electrical_connection_power: 1.1,
		electrical_connection_rotation: 1500,
		created_at: '2026-01-01T00:00:00Z',
		updated_at: '2026-01-01T00:00:00Z',
		...overrides
	};
}

export function buildBacnetObject(overrides: Partial<BacnetObject> = {}): BacnetObject {
	return {
		id: 'bo-1',
		text_fix: 'TF-001',
		description: 'Bacnet object',
		gms_visible: true,
		optional: false,
		text_individual: 'Individual text',
		software_type: 'ai',
		software_number: 1,
		hardware_type: 'di',
		hardware_quantity: 2,
		field_device_id: 'fd-1',
		alarm_type_id: 'alarm-1',
		created_at: '2026-01-01T00:00:00Z',
		updated_at: '2026-01-01T00:00:00Z',
		...overrides
	};
}

export function buildFieldDevice(overrides: Partial<FieldDevice> = {}): FieldDevice {
	return {
		id: 'fd-1',
		bmk: 'FD-001',
		description: 'Field Device',
		text_fix: 'TXT-1',
		apparat_nr: '11',
		sps_controller_system_type_id: 'sps-1',
		system_part_id: 'sp-1',
		specification_id: 'spec-1',
		apparat_id: 'app-1',
		created_at: '2026-01-01T00:00:00Z',
		updated_at: '2026-01-01T00:00:00Z',
		specification: buildSpecification(),
		bacnet_objects: [buildBacnetObject()],
		...overrides
	};
}
