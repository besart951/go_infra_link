import type { Pagination } from '../utils/index.js';

export interface Apparat {
	id: string;
	short_name: string;
	name: string;
	description?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateApparatRequest {
	short_name: string;
	name: string;
	description?: string;
}

export interface UpdateApparatRequest {
	short_name?: string;
	name?: string;
	description?: string;
}

export interface ApparatListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface ApparatListResponse extends Pagination<Apparat> {
	total_pages: number;
}
