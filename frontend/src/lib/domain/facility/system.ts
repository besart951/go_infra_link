import type { Pagination } from '../utils/index.js';

// SystemType
export interface SystemType {
	id: string;
	number_min: number;
	number_max: number;
	name: string;
	created_at: string;
	updated_at: string;
}

export interface CreateSystemTypeRequest {
	number_min: number;
	number_max: number;
	name: string;
}

export interface UpdateSystemTypeRequest {
	number_min?: number;
	number_max?: number;
	name?: string;
}

export interface SystemTypeListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface SystemTypeListResponse extends Pagination<SystemType> {
	total_pages: number; // Backend returns this
}

// SystemPart
export interface SystemPart {
	id: string;
	short_name: string;
	name: string;
	description?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateSystemPartRequest {
	short_name: string;
	name: string;
	description?: string;
}

export interface UpdateSystemPartRequest {
	short_name?: string;
	name?: string;
	description?: string;
}

export interface SystemPartListParams {
	page?: number;
	limit?: number;
	search?: string;
	apparat_id?: string;
}

export interface SystemPartListResponse extends Pagination<SystemPart> {
	total_pages: number;
}
