/**
 * Domain entity representing an SPS Controller System Type
 */
export interface SPSControllerSystemType {
	id: string;
	number?: number | null;
	document_name?: string | null;
	sps_controller_id: string;
	system_type_id: string;
	created_at: string;
	updated_at: string;
}
