import { describe, expect, it, beforeEach, vi } from 'vitest';
import type { SPSController } from '$lib/domain/facility/index.js';
import { encodeMultiFilter } from '$lib/components/facility/shared/projectFacilityListFilters.js';
import { ProjectSPSControllerFetchStrategy } from './ProjectSPSControllerFetchStrategy.js';

const mockListSPSControllers = vi.fn();
const mockGetBulkControllers = vi.fn();
const mockGetBulkCabinets = vi.fn();

vi.mock('$lib/infrastructure/api/projectRepository.js', () => ({
  projectRepository: {
    listSPSControllers: (...args: unknown[]) => mockListSPSControllers(...args)
  }
}));

vi.mock('$lib/infrastructure/api/spsControllerRepository.js', () => ({
  spsControllerRepository: {
    getBulk: (...args: unknown[]) => mockGetBulkControllers(...args)
  }
}));

vi.mock('$lib/infrastructure/api/controlCabinetRepository.js', () => ({
  controlCabinetRepository: {
    getBulk: (...args: unknown[]) => mockGetBulkCabinets(...args)
  }
}));

const controllers: SPSController[] = [
  {
    id: 'controller-1',
    control_cabinet_id: 'cabinet-1',
    device_name: 'SPS 1',
    ga_device: 'GA-1',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  },
  {
    id: 'controller-2',
    control_cabinet_id: 'cabinet-1',
    device_name: 'SPS 2',
    ga_device: 'GA-2',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  },
  {
    id: 'controller-3',
    control_cabinet_id: 'cabinet-2',
    device_name: 'SPS 3',
    ga_device: 'GA-3',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  }
];

describe('ProjectSPSControllerFetchStrategy', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListSPSControllers.mockResolvedValue({
      items: controllers.map((controller) => ({
        id: `link-${controller.id}`,
        project_id: 'project-1',
        sps_controller_id: controller.id
      }))
    });
    mockGetBulkControllers.mockResolvedValue(controllers);
    mockGetBulkCabinets.mockResolvedValue([
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
        building_id: 'building-2',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      }
    ]);
  });

  it('filters project SPS controllers by selected control cabinet ids', async () => {
    const strategy = new ProjectSPSControllerFetchStrategy('project-1');

    const response = await strategy.fetch({
      page: 1,
      pageSize: 10,
      searchText: '',
      filters: { controlCabinetIds: encodeMultiFilter(['cabinet-1']) }
    });

    expect(response.items.map((item) => item.id)).toEqual(['controller-1', 'controller-2']);
    expect(strategy.getCabinetFilterOptions()).toEqual([
      { id: 'cabinet-1', label: 'CC-1', count: 2 },
      { id: 'cabinet-2', label: 'CC-2', count: 1 }
    ]);
  });

  it('ignores stale cabinet filter ids when a project no longer contains them', async () => {
    const strategy = new ProjectSPSControllerFetchStrategy('project-1');

    const response = await strategy.fetch({
      page: 1,
      pageSize: 10,
      searchText: '',
      filters: { controlCabinetIds: encodeMultiFilter(['missing-cabinet']) }
    });

    expect(response.items).toHaveLength(3);
  });
});
