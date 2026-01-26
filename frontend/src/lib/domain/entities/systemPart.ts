/**
 * Domain entity representing a System Part
 */
export interface SystemPart {
	id: string;
	short_name: string;
	name: string;
	description?: string | null;
	created_at: string;
	updated_at: string;
}
