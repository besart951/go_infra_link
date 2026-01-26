/**
 * Domain entity representing a Control Cabinet
 */
export interface ControlCabinet {
	id: string;
	building_id: string;
	control_cabinet_nr?: string | null;
	created_at: string;
	updated_at: string;
}
