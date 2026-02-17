import type { SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest } from '$lib/domain/facility/index.js';
import type { SystemTypeRepository } from '$lib/domain/ports/facility/systemTypeRepository.js';
import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';

export class ListSystemTypesUseCase extends ListEntityUseCase<SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest> {
    constructor(repository: SystemTypeRepository) {
        super(repository);
    }
}
