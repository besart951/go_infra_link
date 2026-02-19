import type { AlarmType, AlarmTypeListResponse } from '$lib/domain/facility/index.js';

export interface AlarmTypeRepository {
    list(params: { search?: string; page?: number; pageSize?: number }, signal?: AbortSignal): Promise<{ items: AlarmType[]; total: number; page: number; totalPages: number }>;
    get(id: string, signal?: AbortSignal): Promise<AlarmType>;
    getWithFields(id: string, signal?: AbortSignal): Promise<AlarmType>;
}
