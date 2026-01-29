import type { Pagination } from '../utils/index.js';
import type { SystemPart } from './system.js';

export interface Apparat {
	id: string;
	short_name: string;
	name: string;
	description?: string;
	system_parts?: SystemPart[];
	created_at: string;
	updated_at: string;
}

export interface CreateApparatRequest {
	short_name: string;
	name: string;
	description?: string;
	system_part_ids?: string[];
}

export interface UpdateApparatRequest {
	short_name?: string;
	name?: string;
	description?: string;
	system_part_ids?: string[];
}

export interface ApparatListParams {
	page?: number;
	limit?: number;
	search?: string;
	object_data_id?: string;
}

export interface ApparatListResponse extends Pagination<Apparat> {
	total_pages: number;
}
