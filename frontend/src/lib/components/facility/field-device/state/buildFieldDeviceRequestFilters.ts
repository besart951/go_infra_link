import type { DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from './types.js';

export function buildFieldDeviceRequestFilters(
  query: DataTableQuery<FieldDeviceFilters>,
  projectId?: string
): Record<string, string> {
  const filters: Record<string, string> = {};

  if (query.orderBy) filters.order_by = query.orderBy;
  if (query.order) filters.order = query.order;
  if (query.filters.buildingId) filters.building_id = query.filters.buildingId;
  if (query.filters.controlCabinetId) {
    filters.control_cabinet_id = query.filters.controlCabinetId;
  }
  if (query.filters.spsControllerId) filters.sps_controller_id = query.filters.spsControllerId;
  if (query.filters.spsControllerSystemTypeId) {
    filters.sps_controller_system_type_id = query.filters.spsControllerSystemTypeId;
  }

  const effectiveProjectId = projectId ?? query.filters.projectId;
  if (effectiveProjectId) {
    filters.project_id = effectiveProjectId;
  }

  return filters;
}
