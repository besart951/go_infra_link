import type { Pagination } from '../utils/index.js';

export interface ObjectData {
	id: string;
	description: string;
	version: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface ObjectDataListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface ObjectDataListResponse extends Pagination<ObjectData> {
	total_pages: number;
}
