import type { Pagination } from '../utils/index.js';

export interface Specification {
	id: string;
	field_device_id?: string;
	specification_supplier?: string;
	specification_brand?: string;
	specification_type?: string;
	additional_info_motor_valve?: string;
	additional_info_size?: number;
	additional_information_installation_location?: string;
	electrical_connection_ph?: number;
	electrical_connection_acdc?: string;
	electrical_connection_amperage?: number;
	electrical_connection_power?: number;
	electrical_connection_rotation?: number;
	created_at: string;
	updated_at: string;
}

export interface CreateSpecificationRequest {
	field_device_id: string;
	specification_supplier?: string;
	// ... add other fields as needed
}

export interface SpecificationListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface SpecificationListResponse extends Pagination<Specification> {
	total_pages: number;
}
