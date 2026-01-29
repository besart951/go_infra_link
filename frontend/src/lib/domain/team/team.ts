/**
 * Team domain types
 * Mirrors backend: internal/domain/team/team.go
 */

import type { Pagination } from '../utils/index.ts';

export interface Team {
	id: string;
	name: string;
	description: string;
	created_at: string;
	updated_at: string;
}

export interface CreateTeamRequest {
	name: string;
	description?: string;
}

export interface UpdateTeamRequest {
	name?: string;
	description?: string;
}

export interface TeamListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface TeamListResponse extends Pagination<Team> {}
