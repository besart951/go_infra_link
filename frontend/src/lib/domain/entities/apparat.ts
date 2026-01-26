/**
 * Domain entity representing an Apparat
 */
export interface Apparat {
	id: string;
	short_name: string;
	name: string;
	description?: string | null;
	created_at: string;
	updated_at: string;
}
