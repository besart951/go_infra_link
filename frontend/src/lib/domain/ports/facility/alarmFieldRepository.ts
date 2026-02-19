import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type {
	AlarmField,
	CreateAlarmFieldRequest,
	UpdateAlarmFieldRequest
} from '$lib/domain/facility/index.js';

export interface AlarmFieldRepository extends CrudRepository<
	AlarmField,
	CreateAlarmFieldRequest,
	UpdateAlarmFieldRequest
> {}
