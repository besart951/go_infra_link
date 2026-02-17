import type { StateText, CreateStateTextRequest, UpdateStateTextRequest } from '$lib/domain/facility/index.js';
import type { StateTextRepository } from '$lib/domain/ports/facility/stateTextRepository.js';

export class ManageStateTextUseCase {
    constructor(private repository: StateTextRepository) { }

    async create(data: CreateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
