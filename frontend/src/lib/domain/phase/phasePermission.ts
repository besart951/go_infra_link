import type { UserRole } from '../user/index.js';

export interface PhasePermission {
  id: string;
  phase_id: string;
  role: UserRole;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

export interface PhasePermissionListResponse {
  items: PhasePermission[];
}

export interface CreatePhasePermissionRequest {
  phase_id: string;
  role: UserRole;
  permissions: string[];
}

export interface UpdatePhasePermissionRequest {
  phase_id?: string;
  role?: UserRole;
  permissions?: string[];
}
