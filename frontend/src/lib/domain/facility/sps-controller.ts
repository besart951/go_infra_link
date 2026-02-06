/**
 * SPS Controller domain types
 * Mirrors backend: internal/domain/facility/sps_controller.go
 */

import type { Pagination } from '../utils/index.ts';

export interface SPSController {
	id: string;
	control_cabinet_id: string;
	ga_device?: string;
	device_name: string;
	device_description?: string;
	device_location?: string;
	ip_address?: string;
	subnet?: string;
	gateway?: string;
	vlan?: string;
	created_at: string;
	updated_at: string;
}

export interface SPSControllerSystemTypeInput {
	system_type_id: string;
	number?: number;
	document_name?: string;
}

export interface CreateSPSControllerRequest {
	ga_device: string;
	device_name: string;
	ip_address: string;
	control_cabinet_id: string;
	system_types?: SPSControllerSystemTypeInput[];
}

export interface UpdateSPSControllerRequest {
	id: string;
	ga_device?: string;
	device_name?: string;
	ip_address?: string;
	control_cabinet_id?: string;
	system_types?: SPSControllerSystemTypeInput[];
}

export interface SPSControllerListParams {
	page?: number;
	limit?: number;
	search?: string;
	control_cabinet_id?: string;
}

export interface SPSControllerListResponse extends Pagination<SPSController> {}
