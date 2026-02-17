import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { SystemPart, CreateSystemPartRequest, UpdateSystemPartRequest } from '$lib/domain/facility/index.js';

export interface SystemPartRepository extends CrudRepository<SystemPart, CreateSystemPartRequest, UpdateSystemPartRequest> {
}
