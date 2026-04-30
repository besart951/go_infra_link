import { api, type ApiOptions } from '$lib/api/client.js';
import type {
  CreatePhasePermissionRequest,
  PhasePermission,
  PhasePermissionListResponse,
  UpdatePhasePermissionRequest
} from '$lib/domain/phase/index.js';

export async function listPhasePermissions(
  phaseId?: string,
  options?: ApiOptions
): Promise<PhasePermissionListResponse> {
  const searchParams = new URLSearchParams();
  if (phaseId) searchParams.set('phase_id', phaseId);
  const query = searchParams.toString();
  return api<PhasePermissionListResponse>(`/phase-permissions${query ? `?${query}` : ''}`, options);
}

export async function createPhasePermission(
  data: CreatePhasePermissionRequest,
  options?: ApiOptions
): Promise<PhasePermission> {
  return api<PhasePermission>('/phase-permissions', {
    ...options,
    method: 'POST',
    body: JSON.stringify(data)
  });
}

export async function updatePhasePermission(
  id: string,
  data: UpdatePhasePermissionRequest,
  options?: ApiOptions
): Promise<PhasePermission> {
  return api<PhasePermission>(`/phase-permissions/${id}`, {
    ...options,
    method: 'PATCH',
    body: JSON.stringify(data)
  });
}

export async function deletePhasePermission(id: string, options?: ApiOptions): Promise<void> {
  return api<void>(`/phase-permissions/${id}`, {
    ...options,
    method: 'DELETE'
  });
}
