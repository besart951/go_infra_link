import type { Pagination } from '../utils/index.js';

export interface SPSControllerSystemType {
	id: string;
	sps_controller_id: string;
	system_type_id: string;
	sps_controller_name?: string;
	system_type_name?: string;
	number?: number;
	document_name?: string;
	created_at: string;
	updated_at: string;
}

export interface SPSControllerSystemTypeListParams {
	page?: number;
	limit?: number;
	search?: string;
	sps_controller_id?: string;
}

export interface SPSControllerSystemTypeListResponse extends Pagination<SPSControllerSystemType> {
	total_pages: number;
}
