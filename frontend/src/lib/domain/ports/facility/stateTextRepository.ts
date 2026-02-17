import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { StateText, CreateStateTextRequest, UpdateStateTextRequest } from '$lib/domain/facility/index.js';

export interface StateTextRepository extends CrudRepository<StateText, CreateStateTextRequest, UpdateStateTextRequest> {
}
