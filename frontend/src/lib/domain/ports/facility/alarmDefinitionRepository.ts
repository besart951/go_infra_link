import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { AlarmDefinition, CreateAlarmDefinitionRequest, UpdateAlarmDefinitionRequest } from '$lib/domain/facility/index.js';

export interface AlarmDefinitionRepository extends CrudRepository<AlarmDefinition, CreateAlarmDefinitionRequest, UpdateAlarmDefinitionRequest> {
}
