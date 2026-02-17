import type { AlarmDefinition, CreateAlarmDefinitionRequest, UpdateAlarmDefinitionRequest } from '$lib/domain/facility/index.js';
import type { AlarmDefinitionRepository } from '$lib/domain/ports/facility/alarmDefinitionRepository.js';

export class ManageAlarmDefinitionUseCase {
    constructor(private repository: AlarmDefinitionRepository) { }

    async create(data: CreateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
