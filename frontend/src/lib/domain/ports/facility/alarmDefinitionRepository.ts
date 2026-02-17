import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { AlarmDefinition, CreateAlarmDefinitionRequest, UpdateAlarmDefinitionRequest } from '$lib/domain/facility/index.js';

export interface AlarmDefinitionRepository extends ListRepository<AlarmDefinition> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<AlarmDefinition>>;
    get(id: string, signal?: AbortSignal): Promise<AlarmDefinition>;
    create(data: CreateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition>;
    update(id: string, data: UpdateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
