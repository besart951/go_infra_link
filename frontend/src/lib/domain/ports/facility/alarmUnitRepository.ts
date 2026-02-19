import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { Unit, CreateUnitRequest, UpdateUnitRequest } from '$lib/domain/facility/index.js';

export interface AlarmUnitRepository extends CrudRepository<
	Unit,
	CreateUnitRequest,
	UpdateUnitRequest
> {}
