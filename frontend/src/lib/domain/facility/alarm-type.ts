/**
 * AlarmType domain types
 * Mirrors backend: internal/domain/facility/alarm_type.go
 */

export interface Unit {
	id: string;
	code: string;
	symbol: string;
	name: string;
}

export interface AlarmField {
	id: string;
	key: string;
	label: string;
	data_type: 'number' | 'integer' | 'boolean' | 'string' | 'enum' | 'duration' | 'state_map' | 'json';
	default_unit_code?: string;
}

export interface AlarmTypeField {
	id: string;
	alarm_type_id: string;
	alarm_field_id: string;
	alarm_field?: AlarmField;
	display_order: number;
	is_required: boolean;
	is_user_editable: boolean;
	default_value_json?: string;
	validation_json?: string;
	default_unit_id?: string;
	default_unit?: Unit;
	ui_group?: string;
	created_at: string;
	updated_at: string;
}

export interface AlarmType {
	id: string;
	code: string;
	name: string;
	fields?: AlarmTypeField[];
	created_at: string;
	updated_at: string;
}

export interface AlarmTypeListResponse {
	items: AlarmType[];
	total: number;
	page: number;
	total_pages: number;
}

export interface AlarmValueDraft {
	alarm_type_field_id: string;
	value_number?: number;
	value_integer?: number;
	value_boolean?: boolean;
	value_string?: string;
	value_json?: string;
	unit_id?: string;
	source?: string;
}

export interface AlarmValue {
	id: string;
	bacnet_object_id: string;
	alarm_type_field_id: string;
	value_number?: number;
	value_integer?: number;
	value_boolean?: boolean;
	value_string?: string;
	value_json?: string;
	unit_id?: string;
	source: string;
	created_at: string;
	updated_at: string;
}

export interface AlarmValuesResponse {
	items: AlarmValue[];
}
