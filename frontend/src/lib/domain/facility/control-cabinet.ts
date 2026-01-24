/**
 * Control Cabinet domain types
 * Mirrors backend: internal/domain/facility/control_cabinet.go
 */

export interface ControlCabinet {
	id: string;
	control_cabinet_nr: string;
	building_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateControlCabinetRequest {
	control_cabinet_nr: string;
	building_id: string;
}

export interface UpdateControlCabinetRequest {
	control_cabinet_nr?: string;
	building_id?: string;
}

export interface ControlCabinetListParams {
	page?: number;
	limit?: number;
	search?: string;
	building_id?: string;
}

export interface ControlCabinetListResponse {
	control_cabinets: ControlCabinet[];
	total: number;
	page: number;
	limit: number;
}
