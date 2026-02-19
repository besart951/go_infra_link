import type { BacnetAlarmRepository } from '$lib/domain/ports/facility/bacnetAlarmRepository.js';
import type { AlarmType, AlarmValue, AlarmValueDraft, AlarmValuesResponse } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const bacnetAlarmRepository: BacnetAlarmRepository = {
    async getSchema(bacnetObjectId: string, signal?: AbortSignal): Promise<AlarmType | null> {
        try {
            const result = await api<AlarmType | null>(`/facility/bacnet-objects/${bacnetObjectId}/alarm-schema`, { signal });
            return result;
        } catch {
            return null;
        }
    },

    async getValues(bacnetObjectId: string, signal?: AbortSignal): Promise<AlarmValue[]> {
        const response = await api<AlarmValuesResponse>(`/facility/bacnet-objects/${bacnetObjectId}/alarm-values`, { signal });
        return response.items;
    },

    async putValues(bacnetObjectId: string, values: AlarmValueDraft[], signal?: AbortSignal): Promise<AlarmValue[]> {
        const response = await api<AlarmValuesResponse>(`/facility/bacnet-objects/${bacnetObjectId}/alarm-values`, {
            method: 'PUT',
            body: JSON.stringify({ values }),
            signal
        });
        return response.items;
    }
};
