import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type {
  ControlCabinet,
  CreateControlCabinetRequest,
  UpdateControlCabinetRequest,
  ControlCabinetDeleteImpact
} from '$lib/domain/facility/index.js';

export interface ControlCabinetRepository extends CrudRepository<
  ControlCabinet,
  CreateControlCabinetRequest,
  UpdateControlCabinetRequest
> {
  getBulk(ids: string[], signal?: AbortSignal): Promise<ControlCabinet[]>;
  copy(
    id: string,
    operationId: string,
    signal?: AbortSignal
  ): Promise<import('$lib/domain/facility/index.js').FacilityJob>;
  validate(
    data: { id?: string; building_id: string; control_cabinet_nr?: string },
    signal?: AbortSignal
  ): Promise<void>;
  getDeleteImpact(id: string, signal?: AbortSignal): Promise<ControlCabinetDeleteImpact>;
}
