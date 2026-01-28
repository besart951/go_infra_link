/**
 * Phase domain types
 * Mirrors backend: internal/domain/project/phase.go
 */

import type { Pagination } from '../utils/index.js';

export interface Phase {
	id: string;
	name: string;
	created_at: string;
	updated_at: string;
}

export interface CreatePhaseRequest {
	name: string;
}

export interface UpdatePhaseRequest {
	name?: string;
}

export interface PhaseListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface PhaseListResponse extends Pagination<Phase> {}
