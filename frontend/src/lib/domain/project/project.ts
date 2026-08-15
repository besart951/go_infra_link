/**
 * Project domain types
 * Mirrors backend: internal/domain/project/project.go
 */

import type { Pagination } from '../utils/index.ts';
import type { User } from '../user/index.ts';
import type { ObjectData } from '../facility/index.ts';
import type { Phase } from '../phase/index.ts';
import type { components } from '$lib/api/generated/schema.js';
export type ProjectStatus = 'planned' | 'ongoing' | 'completed';

export interface Project {
  id: string;
  name: string;
  description: string;
  status: ProjectStatus;
  start_date?: string;
  phase_id: string;
  phase?: Phase | null;
  creator_id: string;
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  status?: ProjectStatus;
  start_date?: string;
  phase_id: string;
}

/** API DTO for PATCH /projects/{id}; the ID belongs exclusively in the path. */
export type UpdateProjectRequest =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_project.UpdateProjectRequest'];

export interface ProjectListParams {
  page?: number;
  limit?: number;
  search?: string;
  status?: ProjectStatus;
  phase_id?: string;
}

export interface ProjectListResponse extends Pagination<Project> {}

export interface ProjectUserListResponse {
  items: User[];
}

export interface ProjectObjectDataListResponse extends Pagination<ObjectData> {}
