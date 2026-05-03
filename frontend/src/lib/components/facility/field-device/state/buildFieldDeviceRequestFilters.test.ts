import { describe, expect, it } from 'vitest';
import { encodeMultiFilter } from '$lib/components/facility/shared/projectFacilityListFilters.js';
import { buildFieldDeviceRequestFilters } from './buildFieldDeviceRequestFilters.js';
import type { DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from './types.js';

function query(filters: FieldDeviceFilters): DataTableQuery<FieldDeviceFilters> {
  return {
    page: 1,
    pageSize: 100,
    searchText: '',
    filters
  };
}

describe('buildFieldDeviceRequestFilters', () => {
  it('maps multi-select field device filters to existing API query keys', () => {
    const buildingIds = encodeMultiFilter(['building-1', 'building-2']);
    const controlCabinetIds = encodeMultiFilter(['cabinet-1', 'cabinet-2']);
    const spsControllerIds = encodeMultiFilter(['controller-1', 'controller-2']);
    const spsControllerSystemTypeIds = encodeMultiFilter(['system-type-1', 'system-type-2']);

    const filters = buildFieldDeviceRequestFilters(
      query({
        buildingIds,
        controlCabinetIds,
        spsControllerIds,
        spsControllerSystemTypeIds
      }),
      'project-1'
    );

    expect(filters).toEqual({
      building_id: buildingIds,
      control_cabinet_id: controlCabinetIds,
      sps_controller_id: spsControllerIds,
      sps_controller_system_type_id: spsControllerSystemTypeIds,
      project_id: 'project-1'
    });
  });

  it('keeps the legacy single-value filters for fixed contexts', () => {
    const filters = buildFieldDeviceRequestFilters(
      query({
        buildingId: 'building-1',
        controlCabinetId: 'cabinet-1',
        spsControllerId: 'controller-1',
        spsControllerSystemTypeId: 'system-type-1',
        projectId: 'project-1'
      })
    );

    expect(filters).toEqual({
      building_id: 'building-1',
      control_cabinet_id: 'cabinet-1',
      sps_controller_id: 'controller-1',
      sps_controller_system_type_id: 'system-type-1',
      project_id: 'project-1'
    });
  });
});
