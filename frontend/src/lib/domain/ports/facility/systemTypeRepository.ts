import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest } from '$lib/domain/facility/index.js';

export interface SystemTypeRepository extends CrudRepository<SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest> {
}
