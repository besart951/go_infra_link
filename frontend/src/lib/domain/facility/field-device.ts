/**
 * Field Device domain types
 * Mirrors backend: internal/domain/facility/field_device.go
 */

import type { Pagination } from "../utils/index.ts";

export interface FieldDevice {
	id: string;
	bmk: string;
	description: string;
	apparat_nr: string;
	sps_controller_system_type_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateFieldDeviceRequest {
	bmk: string;
	description: string;
	apparat_nr: string;
	sps_controller_system_type_id: string;
}

export interface UpdateFieldDeviceRequest {
    id: string;
	bmk?: string;
	description?: string;
	apparat_nr?: string;
	sps_controller_system_type_id?: string;
}

export interface FieldDeviceListParams {
	page?: number;
	limit?: number;
	search?: string;
	sps_controller_system_type_id?: string;
}

export interface FieldDeviceListResponse extends Pagination<FieldDevice> {}
