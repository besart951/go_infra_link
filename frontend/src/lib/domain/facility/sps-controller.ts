/**
 * SPS Controller domain types
 * Mirrors backend: internal/domain/facility/sps_controller.go
 */

export interface SPSController {
	id: string;
	ga_device: string;
	device_name: string;
	ip_address: string;
	control_cabinet_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateSPSControllerRequest {
	ga_device: string;
	device_name: string;
	ip_address: string;
	control_cabinet_id: string;
}

export interface UpdateSPSControllerRequest {
	ga_device?: string;
	device_name?: string;
	ip_address?: string;
	control_cabinet_id?: string;
}

export interface SPSControllerListParams {
	page?: number;
	limit?: number;
	search?: string;
	control_cabinet_id?: string;
}

export interface SPSControllerListResponse {
	sps_controllers: SPSController[];
	total: number;
	page: number;
	limit: number;
}
