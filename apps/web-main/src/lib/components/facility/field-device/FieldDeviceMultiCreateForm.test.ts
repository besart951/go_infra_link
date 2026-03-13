/// <reference types="vitest" />

import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import FieldDeviceMultiCreateForm from './FieldDeviceMultiCreateForm.svelte';

const mockMultiCreate = vi.fn();
const mockGetAvailableNumbers = vi.fn();
const mockGetObjectData = vi.fn();

vi.mock('@i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('@i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: vi.fn()
}));

vi.mock('$lib/infrastructure/api/fieldDeviceRepository.js', () => ({
  fieldDeviceRepository: {
    multiCreate: (...args: unknown[]) => mockMultiCreate(...args),
    getAvailableApparatNumbers: (...args: unknown[]) => mockGetAvailableNumbers(...args)
  }
}));

vi.mock('$lib/infrastructure/api/objectDataRepository.js', () => ({
  objectDataRepository: {
    get: (...args: unknown[]) => mockGetObjectData(...args)
  }
}));

vi.mock('$lib/infrastructure/api/spsControllerRepository.js', () => ({
  spsControllerRepository: {
    listSystemTypes: vi.fn().mockResolvedValue({ items: [] }),
    getSystemType: vi.fn().mockResolvedValue(null)
  }
}));

vi.mock('./multi-create/MultiCreateSelectionSection.svelte', async () => ({
  default: (await import('./__tests__/mocks/MockMultiCreateSelectionSection.svelte')).default
}));

vi.mock('./multi-create/MultiCreateRowsSection.svelte', async () => ({
  default: (await import('./__tests__/mocks/MockMultiCreateRowsSection.svelte')).default
}));

describe('FieldDeviceMultiCreateForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAvailableNumbers.mockResolvedValue({ available: [11, 12] });
    mockGetObjectData.mockResolvedValue({ id: 'obj-1', bacnet_objects: [] });
    mockMultiCreate.mockResolvedValue({
      results: [
        {
          index: 0,
          success: true,
          error: '',
          error_field: '',
          field_device: {
            id: 'fd-created-1',
            apparat_nr: '11',
            sps_controller_system_type_id: 'sps-1',
            system_part_id: 'sp-1',
            apparat_id: 'app-1',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z'
          }
        }
      ],
      total_requests: 1,
      success_count: 1,
      failure_count: 0
    });
  });

  it('renders empty create state initially', () => {
    render(FieldDeviceMultiCreateForm, {});
    expect(screen.queryByTestId('rows-count')).not.toBeInTheDocument();
  });

  it('submits create payload with complete row and selection data', async () => {
    render(FieldDeviceMultiCreateForm, {});

    await fireEvent.click(screen.getByTestId('set-selection'));
    await fireEvent.click(screen.getByTestId('set-preselection'));

    await waitFor(() => {
      expect(mockGetAvailableNumbers).toHaveBeenCalledWith(
        'sps-1',
        'app-1',
        'sp-1',
        expect.any(AbortSignal)
      );
    });

    expect(screen.getByTestId('rows-count')).toHaveTextContent('0');

    await fireEvent.click(screen.getByTestId('add-row'));
    expect(screen.getByTestId('rows-count')).toHaveTextContent('1');

    await fireEvent.click(screen.getByTestId('set-row-values'));
    await fireEvent.click(screen.getByTestId('submit-multi-create'));

    await waitFor(() => {
      expect(mockMultiCreate).toHaveBeenCalledWith(
        {
          field_devices: [
            {
              bmk: 'BMK-NEW',
              description: 'New Description',
              text_fix: 'TextFix-NEW',
              apparat_nr: 11,
              sps_controller_system_type_id: 'sps-1',
              system_part_id: 'sp-1',
              apparat_id: 'app-1',
              object_data_id: 'obj-1'
            }
          ]
        },
        undefined
      );
    });
  });

  it('clears successful rows when backend wraps results in preview', async () => {
    mockMultiCreate.mockResolvedValueOnce({
      preview: {
        results: [
          {
            index: 0,
            success: true,
            error: '',
            error_field: '',
            field_device: {
              id: 'fd-created-preview-1',
              apparat_nr: '11',
              sps_controller_system_type_id: 'sps-1',
              system_part_id: 'sp-1',
              apparat_id: 'app-1',
              created_at: '2026-01-01T00:00:00Z',
              updated_at: '2026-01-01T00:00:00Z'
            }
          }
        ],
        total_requests: 1,
        success_count: 1,
        failure_count: 0
      }
    });

    render(FieldDeviceMultiCreateForm, {});

    await fireEvent.click(screen.getByTestId('set-selection'));
    await fireEvent.click(screen.getByTestId('set-preselection'));

    await waitFor(() => {
      expect(mockGetAvailableNumbers).toHaveBeenCalled();
    });

    await fireEvent.click(screen.getByTestId('add-row'));
    expect(screen.getByTestId('rows-count')).toHaveTextContent('1');

    await fireEvent.click(screen.getByTestId('set-row-values'));
    await fireEvent.click(screen.getByTestId('submit-multi-create'));

    await waitFor(() => {
      expect(screen.getByTestId('rows-count')).toHaveTextContent('0');
    });
  });
});
