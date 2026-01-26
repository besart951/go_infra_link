/**
 * Domain entity representing a Team
 */
export interface Team {
	id: string;
	name: string;
	description?: string | null;
	created_at: string;
	updated_at: string;
}
