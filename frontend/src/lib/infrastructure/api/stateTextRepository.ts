import type { StateTextRepository } from '$lib/domain/ports/facility/stateTextRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  StateText,
  StateTextListResponse,
  CreateStateTextRequest,
  UpdateStateTextRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const stateTextRepository: StateTextRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<StateText>> {
    const response = await api<StateTextListResponse>(
      buildListUrl('/facility/state-texts', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<StateText> {
    return api<StateText>(`/facility/state-texts/${id}`, { signal });
  },

  async create(data: CreateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
    return api<StateText>('/facility/state-texts', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(id: string, data: UpdateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
    return api<StateText>(`/facility/state-texts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/state-texts/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
