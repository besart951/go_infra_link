/**
 * Phase API adapter
 * Infrastructure layer - implements phase data operations via HTTP
 */
import { api, type ApiOptions } from '$lib/api/client.js';
import type {
	Phase,
	PhaseListParams,
	PhaseListResponse,
	CreatePhaseRequest,
	UpdatePhaseRequest
} from '$lib/domain/phase/index.js';
import { createApiAdapter } from '$lib/infrastructure/api/apiListAdapter.js';

export async function listPhases(
	params?: PhaseListParams,
	options?: ApiOptions
): Promise<PhaseListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<PhaseListResponse>(`/phases${query ? `?${query}` : ''}`, options);
}

export async function getPhase(id: string, options?: ApiOptions): Promise<Phase> {
	return api<Phase>(`/phases/${id}`, options);
}

export async function createPhase(data: CreatePhaseRequest, options?: ApiOptions): Promise<Phase> {
	return api<Phase>('/phases', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updatePhase(
	id: string,
	data: UpdatePhaseRequest,
	options?: ApiOptions
): Promise<Phase> {
	return api<Phase>(`/phases/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deletePhase(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/phases/${id}`, { ...options, method: 'DELETE' });
}

export const phaseListRepository = createApiAdapter<Phase>('/phases');
