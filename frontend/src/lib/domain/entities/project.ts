/**
 * Domain entity representing a Project
 */
export interface Project {
	id: string;
	name: string;
	description: string;
	status: 'planned' | 'ongoing' | 'completed';
	start_date?: string | null;
	phase_id: string;
	creator_id: string;
	created_at: string;
	updated_at: string;
}
