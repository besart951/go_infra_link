/**
 * Domain entity representing Object Data
 */
export interface ObjectData {
	id: string;
	description: string;
	obj_version: string;
	is_active: boolean;
	project_id?: string | null;
	created_at: string;
	updated_at: string;
}
