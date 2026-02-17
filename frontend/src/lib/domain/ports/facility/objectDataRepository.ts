import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest } from '$lib/domain/facility/index.js';
import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';

export interface ObjectDataRepository extends CrudRepository<ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest> {
    getBacnetObjects(id: string, signal?: AbortSignal): Promise<BacnetObject[]>;
}
