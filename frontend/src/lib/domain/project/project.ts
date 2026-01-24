/**
 * Project domain types
 * Mirrors backend: internal/domain/project/project.go
 */

export type ProjectStatus = 'planned' | 'ongoing' | 'completed';

export interface Project {
	id: string;
	name: string;
	description: string;
	status: ProjectStatus;
	start_date?: string;
	phase_id: string;
	creator_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateProjectRequest {
	name: string;
	description?: string;
	status?: ProjectStatus;
	start_date?: string;
	phase_id?: string;
}

export interface UpdateProjectRequest {
	name?: string;
	description?: string;
	status?: ProjectStatus;
	start_date?: string;
	phase_id?: string;
}

export interface ProjectListParams {
	page?: number;
	limit?: number;
	search?: string;
	status?: ProjectStatus;
}

export interface ProjectListResponse {
	projects: Project[];
	total: number;
	page: number;
	limit: number;
}
