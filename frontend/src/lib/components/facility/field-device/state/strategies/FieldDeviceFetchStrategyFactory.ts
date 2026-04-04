import type { FieldDevice } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from '../types.js';
import { ContextualFieldDeviceFetchStrategy } from './ContextualFieldDeviceFetchStrategy.js';

export class FieldDeviceFetchStrategyFactory {
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  create(): DataTableFetchStrategy<FieldDevice, FieldDeviceFilters> {
    return new ContextualFieldDeviceFetchStrategy(this.resolveProjectId);
  }
}
