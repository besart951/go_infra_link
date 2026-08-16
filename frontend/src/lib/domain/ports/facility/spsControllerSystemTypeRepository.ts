import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  CopyJob,
  SPSControllerSystemType,
  UpdateSPSControllerSystemTypeRequest
} from '$lib/domain/facility/index.js';

export interface SPSControllerSystemTypeRepository {
  list(
    params: ListParams,
    signal?: AbortSignal
  ): Promise<PaginatedResponse<SPSControllerSystemType>>;
  get(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;
  update(
    id: string,
    data: UpdateSPSControllerSystemTypeRequest,
    signal?: AbortSignal
  ): Promise<SPSControllerSystemType>;
  copy(id: string, operationId: string, signal?: AbortSignal): Promise<CopyJob>;
  delete(id: string, signal?: AbortSignal): Promise<void>;
}
