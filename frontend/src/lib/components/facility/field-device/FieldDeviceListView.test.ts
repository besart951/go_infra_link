/// <reference types="vitest" />

import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { writable, type Writable } from 'svelte/store';
import FieldDeviceListView from './FieldDeviceListView.svelte';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';

const mockDelete = vi.fn();
const mockBulkDelete = vi.fn();
const mockAddToast = vi.fn();
const mockLoad = vi.fn();
const mockReload = vi.fn();
const mockUpdateItem = vi.fn();

type StoreState = {
  items: ReturnType<typeof buildFieldDevice>[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  searchText: string;
  orderBy?: string;
  order?: 'asc' | 'desc';
  filters: Record<string, string>;
  loading: boolean;
  error: string | null;
};

let storeState: Writable<StoreState>;

function resetStore(items: ReturnType<typeof buildFieldDevice>[]) {
  storeState = writable({
    items,
    total: items.length,
    page: 1,
    pageSize: 300,
    totalPages: 1,
    searchText: '',
    orderBy: undefined,
    order: undefined,
    filters: {},
    loading: false,
    error: null
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

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: () => true
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: (...args: unknown[]) => mockAddToast(...args)
}));

vi.mock('$lib/hooks/useUnsavedChangesWarning.svelte.js', () => ({
  useUnsavedChangesWarning: () => undefined
}));

vi.mock('$lib/hooks/useFieldDeviceEditing.svelte.js', () => ({
  useFieldDeviceEditing: () => ({
    hasUnsavedChanges: false,
    pendingCount: 0,
    saveAllPendingEdits: vi.fn(),
    discardAllEdits: vi.fn()
  })
}));

vi.mock('$lib/stores/facility/fieldDeviceStore.js', () => ({
  createFieldDeviceStore: () => ({
    subscribe: storeState.subscribe,
    load: mockLoad,
    reload: mockReload,
    search: vi.fn(),
    setSort: vi.fn(),
    goToPage: vi.fn(),
    setFilters: vi.fn(),
    clearAllFilters: vi.fn(),
    updateItem: mockUpdateItem
  })
}));

vi.mock('$lib/infrastructure/api/fieldDeviceRepository.js', () => ({
  fieldDeviceRepository: {
    delete: (...args: unknown[]) => mockDelete(...args),
    bulkDelete: (...args: unknown[]) => mockBulkDelete(...args),
    multiCreate: vi.fn(),
    bulkUpdate: vi.fn(),
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    getOptions: vi.fn(),
    getOptionsForProject: vi.fn(),
    getAvailableApparatNumbers: vi.fn(),
    createExport: vi.fn(),
    getExportJob: vi.fn(),
    getExportDownloadUrl: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/apparatRepository.js', () => ({
  apparatRepository: {
    list: vi.fn().mockResolvedValue({
      items: [],
      metadata: { total: 0, page: 1, pageSize: 1000, totalPages: 1 }
    })
  }
}));

vi.mock('$lib/infrastructure/api/systemPartRepository.js', () => ({
  systemPartRepository: {
    list: vi.fn().mockResolvedValue({
      items: [],
      metadata: { total: 0, page: 1, pageSize: 1000, totalPages: 1 }
    })
  }
}));

vi.mock('$lib/infrastructure/api/projectRepository.js', () => ({
  projectRepository: {
    addFieldDevice: vi.fn().mockResolvedValue(undefined)
  }
}));

vi.mock('./FieldDeviceSearchBar.svelte', async () => ({
  default: (await import('./__tests__/mocks/MockFieldDeviceSearchBar.svelte')).default
}));

vi.mock('./FieldDeviceTable.svelte', async () => ({
  default: (await import('./__tests__/mocks/MockFieldDeviceTable.svelte')).default
}));

vi.mock('./FieldDeviceMultiCreateForm.svelte', async () => ({
  default: (await import('./__tests__/mocks/MockFieldDeviceMultiCreateForm.svelte')).default
}));

vi.mock('./FieldDeviceFilterCard.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));
vi.mock('./FieldDeviceBulkEditPanel.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));
vi.mock('./FieldDevicePagination.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));
vi.mock('./FieldDeviceFloatingSaveBar.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));
vi.mock('./FieldDeviceExportPanel.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));

describe('FieldDeviceListView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore([]);
    mockDelete.mockResolvedValue(undefined);
    mockBulkDelete.mockResolvedValue({
      results: [{ id: 'fd-1', success: true }],
      total_count: 1,
      success_count: 1,
      failure_count: 0
    });
    vi.mocked(window.confirm).mockReturnValue(true);
  });

  it('renders empty read state and loads on mount', async () => {
    render(FieldDeviceListView, {});

    expect(screen.getByTestId('table-count')).toHaveTextContent('0');
    expect(mockLoad).toHaveBeenCalledTimes(1);
  });

  it('renders populated state when existing devices are provided', async () => {
    resetStore([buildFieldDevice()]);
    render(FieldDeviceListView, {});

    expect(screen.getByTestId('table-count')).toHaveTextContent('1');
  });

  it('calls delete API when single delete is confirmed', async () => {
    const device = buildFieldDevice();
    resetStore([device]);
    render(FieldDeviceListView, {});

    await fireEvent.click(screen.getByTestId(`delete-${device.id}`));

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledTimes(1);
      expect(mockDelete).toHaveBeenCalledWith(device.id, undefined);
      expect(mockReload).toHaveBeenCalled();
    });
  });

  it('does not call delete API when single delete confirmation is rejected', async () => {
    const device = buildFieldDevice();
    resetStore([device]);
    vi.mocked(window.confirm).mockReturnValue(false);
    render(FieldDeviceListView, {});

    await fireEvent.click(screen.getByTestId(`delete-${device.id}`));

    expect(mockDelete).not.toHaveBeenCalled();
  });

  it('calls bulk delete API when selection exists and confirmation is accepted', async () => {
    const device = buildFieldDevice();
    resetStore([device]);
    render(FieldDeviceListView, {});

    await fireEvent.click(screen.getByTestId(`select-${device.id}`));
    await fireEvent.click(screen.getByTestId('bulk-delete'));

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledTimes(1);
      expect(mockBulkDelete).toHaveBeenCalledWith([device.id], undefined);
      expect(mockReload).toHaveBeenCalled();
    });
  });

  it('does not call bulk delete API when confirmation is rejected', async () => {
    const device = buildFieldDevice();
    resetStore([device]);
    vi.mocked(window.confirm).mockReturnValue(false);
    render(FieldDeviceListView, {});

    await fireEvent.click(screen.getByTestId(`select-${device.id}`));
    await fireEvent.click(screen.getByTestId('bulk-delete'));

    expect(mockBulkDelete).not.toHaveBeenCalled();
  });

  it('opens multi-create and reloads after create success callback', async () => {
    render(FieldDeviceListView, {});

    await fireEvent.click(screen.getByText('field_device.actions.multi_create'));
    await fireEvent.click(screen.getByTestId('multi-create-success'));

    await waitFor(() => {
      expect(mockReload).toHaveBeenCalled();
    });
  });
});
