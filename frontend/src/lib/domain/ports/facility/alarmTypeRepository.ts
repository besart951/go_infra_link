import type {
  AlarmType,
  AlarmTypeField,
  CreateAlarmTypeFieldRequest,
  CreateAlarmTypeRequest,
  UpdateAlarmTypeFieldRequest,
  UpdateAlarmTypeRequest
} from '$lib/domain/facility/index.js';
import type { VersionedDeleteCommand } from '$lib/domain/ports/crudRepository.js';

export interface AlarmTypeRepository {
  list(
    params: { search?: string; page?: number; pageSize?: number },
    signal?: AbortSignal
  ): Promise<{ items: AlarmType[]; total: number; page: number; totalPages: number }>;
  get(id: string, signal?: AbortSignal): Promise<AlarmType>;
  getWithFields(id: string, signal?: AbortSignal): Promise<AlarmType>;
  create(data: CreateAlarmTypeRequest, signal?: AbortSignal): Promise<AlarmType>;
  update(id: string, data: UpdateAlarmTypeRequest, signal?: AbortSignal): Promise<AlarmType>;
  delete(command: VersionedDeleteCommand, signal?: AbortSignal): Promise<void>;
  createField(
    alarmTypeId: string,
    data: CreateAlarmTypeFieldRequest,
    signal?: AbortSignal
  ): Promise<AlarmTypeField>;
  updateField(
    id: string,
    data: UpdateAlarmTypeFieldRequest,
    signal?: AbortSignal
  ): Promise<AlarmTypeField>;
  deleteField(command: VersionedDeleteCommand, signal?: AbortSignal): Promise<void>;
}
