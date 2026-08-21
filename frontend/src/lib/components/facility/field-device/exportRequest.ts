import type { CreateFieldDeviceExportRequest } from '$lib/domain/facility/index.js';
import type { FieldDeviceFilters } from './state/types.js';

interface BuildExportRequestInput {
  projectId?: string;
  filters?: FieldDeviceFilters;
  search?: string;
  selectedProjectIds?: string[];
  selectedBuildingIds?: string[];
  selectedControlCabinetIds?: string[];
  selectedSPSControllerIds?: string[];
}

export function buildFieldDeviceExportRequest(
  input: BuildExportRequestInput
): CreateFieldDeviceExportRequest {
  const filters = input.filters ?? {};
  return compactRequest({
    project_ids: unique([
      ...(input.projectId ? [input.projectId] : []),
      ...split(filters.projectId),
      ...(input.selectedProjectIds ?? [])
    ]),
    buildings_id: unique([
      ...split(filters.buildingId),
      ...split(filters.buildingIds),
      ...(input.selectedBuildingIds ?? [])
    ]),
    control_cabinet_id: unique([
      ...split(filters.controlCabinetId),
      ...split(filters.controlCabinetIds),
      ...(input.selectedControlCabinetIds ?? [])
    ]),
    sps_controller_id: unique([
      ...split(filters.spsControllerId),
      ...split(filters.spsControllerIds),
      ...(input.selectedSPSControllerIds ?? [])
    ]),
    sps_controller_system_type_ids: unique([
      ...split(filters.spsControllerSystemTypeId),
      ...split(filters.spsControllerSystemTypeIds)
    ]),
    search: input.search?.trim() || undefined
  });
}

export function hasExportScope(request: CreateFieldDeviceExportRequest): boolean {
  return Boolean(
    request.export_all ||
    request.search ||
    request.project_ids?.length ||
    request.buildings_id?.length ||
    request.control_cabinet_id?.length ||
    request.sps_controller_id?.length ||
    request.sps_controller_system_type_ids?.length
  );
}

function split(value?: string): string[] {
  return value
    ? value
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean)
    : [];
}

function unique(values: string[]): string[] | undefined {
  const result = [...new Set(values.filter(Boolean))];
  return result.length > 0 ? result : undefined;
}

function compactRequest(request: CreateFieldDeviceExportRequest): CreateFieldDeviceExportRequest {
  return Object.fromEntries(
    Object.entries(request).filter(([, value]) => value !== undefined)
  ) as CreateFieldDeviceExportRequest;
}
