/**
 * BACnet Object domain types
 * Mirrors backend: internal/domain/facility/bacnet_object.go
 */

export interface BacnetObject {
	id: string;
	text_fix: string;
	description?: string;
	gms_visible: boolean;
	optional: boolean;
	text_individual?: string;
	software_type: string;
	software_number: number;
	hardware_type: string;
	hardware_quantity: number;
	field_device_id?: string;
	software_reference_id?: string;
	state_text_id?: string;
	notification_class_id?: string;
	alarm_definition_id?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateBacnetObjectRequest {
	text_fix: string;
	description?: string;
	gms_visible: boolean;
	optional: boolean;
	text_individual?: string;
	software_type: string;
	software_number: number;
	hardware_type: string;
	hardware_quantity: number;
	field_device_id?: string;
	software_reference_id?: string;
	state_text_id?: string;
	notification_class_id?: string;
	alarm_definition_id?: string;
}

export interface UpdateBacnetObjectRequest {
	text_fix?: string;
	description?: string;
	gms_visible?: boolean;
	optional?: boolean;
	text_individual?: string;
	software_type?: string;
	software_number?: number;
	hardware_type?: string;
	hardware_quantity?: number;
	field_device_id?: string;
	software_reference_id?: string;
	state_text_id?: string;
	notification_class_id?: string;
	alarm_definition_id?: string;
}

// BACnet Software Types
export const BACNET_SOFTWARE_TYPES = [
	{ value: 'ai', label: 'AI - Analog Input' },
	{ value: 'ao', label: 'AO - Analog Output' },
	{ value: 'av', label: 'AV - Analog Value' },
	{ value: 'bi', label: 'BI - Binary Input' },
	{ value: 'bo', label: 'BO - Binary Output' },
	{ value: 'bv', label: 'BV - Binary Value' },
	{ value: 'mi', label: 'MI - Multi-state Input' },
	{ value: 'mo', label: 'MO - Multi-state Output' },
	{ value: 'mv', label: 'MV - Multi-state Value' },
	{ value: 'ca', label: 'CA - Calendar' },
	{ value: 'ee', label: 'EE - Event Enrollment' },
	{ value: 'lp', label: 'LP - Life Safety Point' },
	{ value: 'nc', label: 'NC - Notification Class' },
	{ value: 'sc', label: 'SC - Schedule' },
	{ value: 'tl', label: 'TL - Trend Log' }
] as const;

// BACnet Hardware Types
export const BACNET_HARDWARE_TYPES = [
	{ value: 'do', label: 'DO - Digital Output' },
	{ value: 'ao', label: 'AO - Analog Output' },
	{ value: 'di', label: 'DI - Digital Input' },
	{ value: 'ai', label: 'AI - Analog Input' }
] as const;
