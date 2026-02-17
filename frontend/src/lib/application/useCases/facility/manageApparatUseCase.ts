import type { Apparat, CreateApparatRequest, UpdateApparatRequest } from '$lib/domain/facility/index.js';
import type { ApparatRepository } from '$lib/domain/ports/facility/apparatRepository.js';

export class ManageApparatUseCase {
    constructor(private repository: ApparatRepository) { }

    async create(data: CreateApparatRequest, signal?: AbortSignal): Promise<Apparat> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateApparatRequest, signal?: AbortSignal): Promise<Apparat> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
