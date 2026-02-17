import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { Apparat, CreateApparatRequest, UpdateApparatRequest } from '$lib/domain/facility/index.js';

export interface ApparatRepository extends CrudRepository<Apparat, CreateApparatRequest, UpdateApparatRequest> {
    getBulk(ids: string[], signal?: AbortSignal): Promise<Apparat[]>;
}
