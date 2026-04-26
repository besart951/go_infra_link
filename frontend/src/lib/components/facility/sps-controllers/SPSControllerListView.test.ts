/// <reference types="vitest" />

import { render, waitFor } from '@testing-library/svelte';
import SPSControllerListView from './SPSControllerListView.svelte';

const mockProjectListSPSControllers = vi.fn();
const mockListControllers = vi.fn();
const mockGetBulkControllers = vi.fn();
const mockGetController = vi.fn();
const mockGetBulkCabinets = vi.fn();
const mockListSystemTypes = vi.fn();

type TestSPSController = {
  id: string;
  control_cabinet_id: string;
  device_name: string;
  ga_device?: string;
  created_at: string;
  updated_at: string;
};

let controllers: TestSPSController[] = [];

function resetControllers(items: TestSPSController[]) {
  controllers = items;
  mockListControllers.mockResolvedValue({
    items,
    metadata: {
      total: items.length,
      page: 1,
      pageSize: 10,
      totalPages: items.length > 0 ? 1 : 0
    }
  });
  mockGetBulkControllers.mockResolvedValue(items);
  mockGetController.mockImplementation(async (id: string) =>
    controllers.find((item) => item.id === id)
  );
  mockGetBulkCabinets.mockResolvedValue(
    [...new Set(items.map((item) => item.control_cabinet_id))].map((id) => ({
      id,
      control_cabinet_nr: `CC-${id}`,
      building_id: 'building-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    }))
  );
  mockListSystemTypes.mockResolvedValue({
    items: items.map((item) => ({
      id: `system-type-${item.id}`,
      sps_controller_id: item.id,
      system_type_id: `type-${item.id}`,
      system_type_name: `Type ${item.id}`,
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    })),
    metadata: {
      total: items.length,
      page: 1,
      pageSize: 1000,
      totalPages: items.length > 0 ? 1 : 0
    }
  });
}

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/infrastructure/api/projectRepository.js', () => ({
  projectRepository: {
    listSPSControllers: (...args: unknown[]) => mockProjectListSPSControllers(...args),
    addSPSController: vi.fn().mockResolvedValue(undefined),
    removeSPSController: vi.fn().mockResolvedValue(undefined),
    copySPSController: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/spsControllerRepository.js', () => ({
  spsControllerRepository: {
    list: (...args: unknown[]) => mockListControllers(...args),
    getBulk: (...args: unknown[]) => mockGetBulkControllers(...args),
    get: (...args: unknown[]) => mockGetController(...args),
    copy: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    validate: vi.fn(),
    getNextGADevice: vi.fn(),
    listSystemTypes: vi.fn(),
    getSystemType: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/controlCabinetRepository.js', () => ({
  controlCabinetRepository: {
    getBulk: (...args: unknown[]) => mockGetBulkCabinets(...args)
  }
}));

vi.mock('$lib/infrastructure/api/spsControllerSystemTypeRepository.js', () => ({
  spsControllerSystemTypeRepository: {
    list: (...args: unknown[]) => mockListSystemTypes(...args)
  }
}));

vi.mock('./SPSControllerList.svelte', async () => ({
  default: (await import('../field-device/__tests__/mocks/Noop.svelte')).default
}));

describe('SPSControllerListView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetControllers([]);
  });

  it('refreshes visible controller ids without full project list reload', async () => {
    const controller = {
      id: 'controller-1',
      control_cabinet_id: 'cabinet-1',
      device_name: 'SPS 1',
      ga_device: 'GA-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetControllers([controller]);
    mockGetController.mockResolvedValue({ ...controller, device_name: 'SPS 1 updated' });

    const rendered = render(SPSControllerListView, {});

    await waitFor(() => {
      expect(mockListControllers).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        entityIds: [controller.id]
      }
    });

    await waitFor(() => {
      expect(mockGetController).toHaveBeenCalledWith(controller.id);
    });
    expect(mockListControllers).toHaveBeenCalledTimes(1);
  });

  it('applies controller delta without follow-up fetch', async () => {
    const controller = {
      id: 'controller-1',
      control_cabinet_id: 'cabinet-1',
      device_name: 'SPS 1',
      ga_device: 'GA-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetControllers([controller]);

    const rendered = render(SPSControllerListView, {});

    await waitFor(() => {
      expect(mockListControllers).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      deltaRequest: {
        key: 1,
        items: [{ ...controller, device_name: 'SPS 1 updated' }]
      }
    });

    await waitFor(() => {
      expect(mockGetController).not.toHaveBeenCalled();
    });
    expect(mockListControllers).toHaveBeenCalledTimes(1);
  });

  it('refreshes cabinet labels without full controller reload', async () => {
    const controller = {
      id: 'controller-1',
      control_cabinet_id: 'cabinet-1',
      device_name: 'SPS 1',
      ga_device: 'GA-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetControllers([controller]);

    const rendered = render(SPSControllerListView, {});

    await waitFor(() => {
      expect(mockListControllers).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      controlCabinetLabelRefreshRequest: {
        key: 1,
        entityIds: [controller.control_cabinet_id]
      }
    });

    await waitFor(() => {
      expect(mockGetBulkCabinets).toHaveBeenCalledTimes(2);
    });
    expect(mockListControllers).toHaveBeenCalledTimes(1);
    expect(mockGetController).not.toHaveBeenCalled();
  });

  it('applies cabinet label delta without follow-up fetch', async () => {
    const controller = {
      id: 'controller-1',
      control_cabinet_id: 'cabinet-1',
      device_name: 'SPS 1',
      ga_device: 'GA-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetControllers([controller]);

    const rendered = render(SPSControllerListView, {});

    await waitFor(() => {
      expect(mockGetBulkCabinets).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      controlCabinetLabelDeltaRequest: {
        key: 1,
        items: [
          {
            id: 'cabinet-1',
            control_cabinet_nr: 'CC-1-updated',
            building_id: 'building-1',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z'
          }
        ]
      }
    });

    await waitFor(() => {
      expect(mockGetBulkCabinets).toHaveBeenCalledTimes(1);
    });
    expect(mockGetController).not.toHaveBeenCalled();
  });
});
