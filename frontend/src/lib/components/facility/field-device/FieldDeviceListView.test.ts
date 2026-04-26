/// <reference types="vitest" />

import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import FieldDeviceListView from './FieldDeviceListView.svelte';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';

const mockDelete = vi.fn();
const mockBulkDelete = vi.fn();
const mockGet = vi.fn();
const mockList = vi.fn();
const mockAddToast = vi.fn();

let listItems: ReturnType<typeof buildFieldDevice>[] = [];

function resetDevices(items: ReturnType<typeof buildFieldDevice>[]) {
  listItems = items;
  mockList.mockImplementation(async () => ({
    items: listItems,
    metadata: {
      total: listItems.length,
      page: 1,
      pageSize: 300,
      totalPages: listItems.length > 0 ? 1 : 0
    }
  }));
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

vi.mock('$lib/infrastructure/api/fieldDeviceRepository.js', () => ({
  fieldDeviceRepository: {
    delete: (...args: unknown[]) => mockDelete(...args),
    bulkDelete: (...args: unknown[]) => mockBulkDelete(...args),
    multiCreate: vi.fn(),
    bulkUpdate: vi.fn(),
    list: (...args: unknown[]) => mockList(...args),
    get: (...args: unknown[]) => mockGet(...args),
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
    addFieldDevice: vi.fn().mockResolvedValue(undefined),
    addFieldDevices: vi.fn().mockResolvedValue({
      success_field_device_ids: [],
      association_errors: []
    })
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
    resetDevices([]);
    mockDelete.mockResolvedValue(undefined);
    mockGet.mockImplementation(async (id: string) => listItems.find((item) => item.id === id));
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
    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });
  });

  it('renders populated state when existing devices are provided', async () => {
    resetDevices([buildFieldDevice()]);
    render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(screen.getByTestId('table-count')).toHaveTextContent('1');
    });
  });

  it('calls delete API when single delete is confirmed', async () => {
    const device = buildFieldDevice();
    resetDevices([device]);
    render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(screen.getByTestId(`delete-${device.id}`)).toBeInTheDocument();
    });

    await fireEvent.click(screen.getByTestId(`delete-${device.id}`));

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledTimes(1);
      expect(mockDelete).toHaveBeenCalledWith(device.id, undefined);
      expect(mockList).toHaveBeenCalledTimes(2);
    });
  });

  it('does not call delete API when single delete confirmation is rejected', async () => {
    const device = buildFieldDevice();
    resetDevices([device]);
    vi.mocked(window.confirm).mockReturnValue(false);
    render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(screen.getByTestId(`delete-${device.id}`)).toBeInTheDocument();
    });

    await fireEvent.click(screen.getByTestId(`delete-${device.id}`));

    expect(mockDelete).not.toHaveBeenCalled();
  });

  it('calls bulk delete API when selection exists and confirmation is accepted', async () => {
    const device = buildFieldDevice();
    resetDevices([device]);
    render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(screen.getByTestId(`select-${device.id}`)).toBeInTheDocument();
    });

    await fireEvent.click(screen.getByTestId(`select-${device.id}`));
    await fireEvent.click(screen.getByTestId('bulk-delete'));

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledTimes(1);
      expect(mockBulkDelete).toHaveBeenCalledWith([device.id], undefined);
      expect(mockList).toHaveBeenCalledTimes(2);
    });
  });

  it('does not call bulk delete API when confirmation is rejected', async () => {
    const device = buildFieldDevice();
    resetDevices([device]);
    vi.mocked(window.confirm).mockReturnValue(false);
    render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(screen.getByTestId(`select-${device.id}`)).toBeInTheDocument();
    });

    await fireEvent.click(screen.getByTestId(`select-${device.id}`));
    await fireEvent.click(screen.getByTestId('bulk-delete'));

    expect(mockBulkDelete).not.toHaveBeenCalled();
  });

  it('opens multi-create and reloads after multi-create success callback', async () => {
    render(FieldDeviceListView, {});

    await fireEvent.click(
      screen.getByRole('button', { name: 'field_device.actions.multi_create' })
    );
    await fireEvent.click(screen.getByTestId('multi-create-success'));

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(2);
    });
  });

  it('refreshes visible device ids without full list reload', async () => {
    const device = buildFieldDevice();
    const updatedDevice = { ...device, description: 'Updated remotely' };
    resetDevices([device]);
    mockGet.mockResolvedValue(updatedDevice);

    const rendered = render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        deviceIds: [device.id]
      }
    });

    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledWith(device.id);
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });

  it('applies field device delta without follow-up fetch', async () => {
    const device = buildFieldDevice();
    const updatedDevice = { ...device, description: 'Updated by delta' };
    resetDevices([device]);

    const rendered = render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        devices: [updatedDevice]
      }
    });

    await waitFor(() => {
      expect(mockGet).not.toHaveBeenCalled();
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });

  it('refreshes visible devices for targeted sps controller ids without full list reload', async () => {
    const device = buildFieldDevice({
      sps_controller_system_type: {
        id: 'system-type-1',
        sps_controller_id: 'controller-1',
        system_type_id: 'type-1',
        sps_controller_name: 'SPS 1',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      }
    });
    const updatedDevice = {
      ...device,
      sps_controller_system_type: {
        ...device.sps_controller_system_type!,
        sps_controller_name: 'SPS 1 updated'
      }
    };
    resetDevices([device]);
    mockGet.mockResolvedValue(updatedDevice);

    const rendered = render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        spsControllerIds: ['controller-1']
      }
    });

    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledWith(device.id);
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });

  it('applies sps controller delta to visible field devices without follow-up fetch', async () => {
    const device = buildFieldDevice({
      sps_controller_system_type: {
        id: 'system-type-1',
        sps_controller_id: 'controller-1',
        system_type_id: 'type-1',
        sps_controller_name: 'SPS 1',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      }
    });
    resetDevices([device]);

    const rendered = render(FieldDeviceListView, {});

    await waitFor(() => {
      expect(mockList).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      refreshRequest: {
        key: 1,
        spsControllers: [
          {
            id: 'controller-1',
            control_cabinet_id: 'cabinet-1',
            device_name: 'SPS 1 updated',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z'
          }
        ]
      }
    });

    await waitFor(() => {
      expect(mockGet).not.toHaveBeenCalled();
    });
    expect(mockList).toHaveBeenCalledTimes(1);
  });
});
