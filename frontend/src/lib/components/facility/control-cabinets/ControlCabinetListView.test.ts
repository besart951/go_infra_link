/// <reference types="vitest" />

import { render, waitFor } from '@testing-library/svelte';
import ControlCabinetListView from './ControlCabinetListView.svelte';

const mockProjectListControlCabinets = vi.fn();
const mockList = vi.fn();
const mockGetBulk = vi.fn();
const mockGet = vi.fn();
const mockGetBuildings = vi.fn();

type TestControlCabinet = {
  id: string;
  control_cabinet_nr: string;
  building_id: string;
  created_at: string;
  updated_at: string;
};

let cabinets: TestControlCabinet[] = [];

function resetCabinets(items: TestControlCabinet[]) {
  cabinets = items;
  mockList.mockResolvedValue({
    items,
    metadata: {
      total: items.length,
      page: 1,
      pageSize: 10,
      totalPages: items.length > 0 ? 1 : 0
    }
  });
  mockGetBulk.mockResolvedValue(items);
  mockGet.mockImplementation(async (id: string) => cabinets.find((item) => item.id === id));
  mockGetBuildings.mockResolvedValue(
    [...new Set(items.map((item) => item.building_id))].map((id) => ({
      id,
      iws_code: 'IWS',
      building_group: id.toUpperCase()
    }))
  );
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
    listControlCabinets: (...args: unknown[]) => mockProjectListControlCabinets(...args),
    addControlCabinet: vi.fn().mockResolvedValue(undefined),
    removeControlCabinet: vi.fn().mockResolvedValue(undefined),
    copyControlCabinet: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/controlCabinetRepository.js', () => ({
  controlCabinetRepository: {
    list: (...args: unknown[]) => mockList(...args),
    getBulk: (...args: unknown[]) => mockGetBulk(...args),
    get: (...args: unknown[]) => mockGet(...args),
    copy: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    validate: vi.fn(),
    getDeleteImpact: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/buildingRepository.js', () => ({
  buildingRepository: {
    getBulk: (...args: unknown[]) => mockGetBuildings(...args)
  }
}));

vi.mock('./ControlCabinetList.svelte', async () => ({
  default: (await import('../field-device/__tests__/mocks/Noop.svelte')).default
}));

describe('ControlCabinetListView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetCabinets([]);
  });

  it('refreshes visible cabinet ids without full project list reload', async () => {
    const cabinet = {
      id: 'cabinet-1',
      control_cabinet_nr: 'CC-1',
      building_id: 'building-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetCabinets([cabinet]);
    mockGet.mockResolvedValue({ ...cabinet, control_cabinet_nr: 'CC-1-updated' });

    const rendered = render(ControlCabinetListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        entityIds: [cabinet.id]
      }
    });

    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledWith(cabinet.id);
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });

  it('applies cabinet delta without follow-up fetch', async () => {
    const cabinet = {
      id: 'cabinet-1',
      control_cabinet_nr: 'CC-1',
      building_id: 'building-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    };

    resetCabinets([cabinet]);

    const rendered = render(ControlCabinetListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      deltaRequest: {
        key: 1,
        items: [{ ...cabinet, control_cabinet_nr: 'CC-1-updated' }]
      }
    });

    await waitFor(() => {
      expect(mockGet).not.toHaveBeenCalled();
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });
});
