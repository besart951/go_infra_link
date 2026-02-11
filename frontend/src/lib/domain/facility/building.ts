/**
 * Building domain types
 * Mirrors backend: internal/domain/facility/building.go
 */

import type { Pagination } from '../utils/index.js';

export interface Building {
	id: string;
	iws_code: string;
	building_group: number;
	created_at: string;
	updated_at: string;
}

export interface CreateBuildingRequest {
	iws_code: string;
	building_group: number;
}

export interface UpdateBuildingRequest {
	iws_code?: string;
	building_group?: number;
}

export interface BuildingListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface BuildingListResponse extends Pagination<Building> {}

export interface BuildingBulkRequest {
	ids: string[];
}

export interface BuildingBulkResponse {
	items: Building[];
}
