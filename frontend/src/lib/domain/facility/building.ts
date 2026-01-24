/**
 * Building domain types
 * Mirrors backend: internal/domain/facility/building.go
 */

export interface Building {
	id: string;
	iws_code: string;
	building_group: string;
	created_at: string;
	updated_at: string;
}

export interface CreateBuildingRequest {
	iws_code: string;
	building_group: string;
}

export interface UpdateBuildingRequest {
	iws_code?: string;
	building_group?: string;
}

export interface BuildingListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface BuildingListResponse {
	buildings: Building[];
	total: number;
	page: number;
	limit: number;
}
