import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest } from '$lib/domain/facility/index.js';
import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';

export interface ObjectDataRepository extends ListRepository<ObjectData> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<ObjectData>>;
    get(id: string, signal?: AbortSignal): Promise<ObjectData>;
    create(data: CreateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData>;
    update(id: string, data: UpdateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
    getBacnetObjects(id: string, signal?: AbortSignal): Promise<BacnetObject[]>;
}
