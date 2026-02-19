import type { AlarmType, AlarmValue, AlarmValueDraft } from '$lib/domain/facility/index.js';

export interface BacnetAlarmRepository {
    getSchema(bacnetObjectId: string, signal?: AbortSignal): Promise<AlarmType | null>;
    getValues(bacnetObjectId: string, signal?: AbortSignal): Promise<AlarmValue[]>;
    putValues(bacnetObjectId: string, values: AlarmValueDraft[], signal?: AbortSignal): Promise<AlarmValue[]>;
}
