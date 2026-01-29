/**
 * Control Cabinet domain types
 * Mirrors backend: internal/domain/facility/control_cabinet.go
 */

import type { Pagination } from '../utils/index.ts';

export interface ControlCabinet {
	id: string;
	control_cabinet_nr: string;
	building_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateControlCabinetRequest {
	control_cabinet_nr: string;
	building_id: string;
}

export interface UpdateControlCabinetRequest {
	id: string;
	control_cabinet_nr?: string;
	building_id?: string;
}

export interface ControlCabinetListParams {
	page?: number;
	limit?: number;
	search?: string;
	building_id?: string;
}

export interface ControlCabinetListResponse extends Pagination<ControlCabinet> {}

export interface ControlCabinetDeleteImpact {
	control_cabinet_id: string;
	sps_controllers_count: number;
	sps_controller_system_types_count: number;
	field_devices_count: number;
	bacnet_objects_count: number;
	specifications_count: number;
}
