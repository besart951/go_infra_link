import { describe, expect, it, beforeEach, vi } from 'vitest';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import { encodeMultiFilter } from '$lib/components/facility/shared/projectFacilityListFilters.js';
import { ProjectControlCabinetFetchStrategy } from './ProjectControlCabinetFetchStrategy.js';

const mockListControlCabinets = vi.fn();
const mockGetBulkCabinets = vi.fn();
const mockGetBulkBuildings = vi.fn();

vi.mock('$lib/infrastructure/api/projectRepository.js', () => ({
  projectRepository: {
    listControlCabinets: (...args: unknown[]) => mockListControlCabinets(...args)
  }
}));

vi.mock('$lib/infrastructure/api/controlCabinetRepository.js', () => ({
  controlCabinetRepository: {
    getBulk: (...args: unknown[]) => mockGetBulkCabinets(...args)
  }
}));

vi.mock('$lib/infrastructure/api/buildingRepository.js', () => ({
  buildingRepository: {
    getBulk: (...args: unknown[]) => mockGetBulkBuildings(...args)
  }
}));

const cabinets: ControlCabinet[] = [
  {
    id: 'cabinet-1',
    control_cabinet_nr: 'CC-1',
    building_id: 'building-1',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  },
  {
    id: 'cabinet-2',
    control_cabinet_nr: 'CC-2',
    building_id: 'building-1',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  },
  {
    id: 'cabinet-3',
    control_cabinet_nr: 'CC-3',
    building_id: 'building-2',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  }
];

describe('ProjectControlCabinetFetchStrategy', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListControlCabinets.mockResolvedValue({
      items: cabinets.map((cabinet) => ({
        id: `link-${cabinet.id}`,
        project_id: 'project-1',
        control_cabinet_id: cabinet.id
      }))
    });
    mockGetBulkCabinets.mockResolvedValue(cabinets);
    mockGetBulkBuildings.mockResolvedValue([
      { id: 'building-1', iws_code: 'IWS', building_group: 'A' },
      { id: 'building-2', iws_code: 'IWS', building_group: 'B' }
    ]);
  });

  it('filters project control cabinets by selected building ids', async () => {
    const strategy = new ProjectControlCabinetFetchStrategy('project-1');

    const response = await strategy.fetch({
      page: 1,
      pageSize: 10,
      searchText: '',
      filters: { buildingIds: encodeMultiFilter(['building-1']) }
    });

    expect(response.items.map((item) => item.id)).toEqual(['cabinet-1', 'cabinet-2']);
    expect(strategy.getBuildingFilterOptions()).toEqual([
      { id: 'building-1', label: 'IWS-A', count: 2 },
      { id: 'building-2', label: 'IWS-B', count: 1 }
    ]);
  });

  it('ignores stale building filter ids when a project no longer contains them', async () => {
    const strategy = new ProjectControlCabinetFetchStrategy('project-1');

    const response = await strategy.fetch({
      page: 1,
      pageSize: 10,
      searchText: '',
      filters: { buildingIds: encodeMultiFilter(['missing-building']) }
    });

    expect(response.items).toHaveLength(3);
  });
});
