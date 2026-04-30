/// <reference types="vitest" />

import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import MultiCreateStateHarness from '../__tests__/MultiCreateStateHarness.svelte';
import { STORAGE_KEY } from '$lib/domain/facility/fieldDeviceMultiCreate.js';

const mockGetAvailableNumbers = vi.fn();
const mockGetObjectData = vi.fn();

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: vi.fn()
}));

vi.mock('$lib/infrastructure/api/fieldDeviceRepository.js', () => ({
  fieldDeviceRepository: {
    getAvailableApparatNumbers: (...args: unknown[]) => mockGetAvailableNumbers(...args),
    multiCreate: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/objectDataRepository.js', () => ({
  objectDataRepository: {
    get: (...args: unknown[]) => mockGetObjectData(...args)
  }
}));

vi.mock('$lib/infrastructure/api/projectRepository.js', () => ({
  projectRepository: {
    createFieldDevices: vi.fn()
  }
}));

vi.mock('$lib/infrastructure/api/spsControllerRepository.js', () => ({
  spsControllerRepository: {
    listSystemTypes: vi.fn().mockResolvedValue({ items: [] }),
    getSystemType: vi.fn().mockResolvedValue(null)
  }
}));

describe('FieldDeviceMultiCreateState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    mockGetAvailableNumbers.mockResolvedValue({ available: [11, 12] });
    mockGetObjectData.mockResolvedValue({ id: 'obj-1', bacnet_objects: [] });
  });

  it('restores persisted selection and rows', () => {
    sessionStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        spsControllerSystemTypeId: 'sps-persisted',
        objectDataId: 'obj-persisted',
        apparatId: 'app-persisted',
        systemPartId: 'sp-persisted',
        rows: [
          {
            id: 'row-1',
            bmk: 'BMK',
            description: 'Description',
            textFix: 'Text',
            apparatNr: 12
          }
        ]
      })
    );

    render(MultiCreateStateHarness);

    expect(screen.getByTestId('selection')).toHaveTextContent(
      'sps-persisted|obj-persisted|app-persisted|sp-persisted'
    );
    expect(screen.getByTestId('rows-count')).toHaveTextContent('1');
  });

  it('aborts stale available-number loading and keeps latest result', async () => {
    const requests: Array<{
      resolve: (value: { available: number[] }) => void;
      signal: AbortSignal;
    }> = [];
    mockGetAvailableNumbers.mockImplementation(
      (_spsId: string, _apparatId: string, _systemPartId: string, signal: AbortSignal) =>
        new Promise((resolve) => {
          requests.push({ resolve, signal });
        })
    );

    render(MultiCreateStateHarness);

    await fireEvent.click(screen.getByTestId('set-sps-1'));
    await fireEvent.click(screen.getByTestId('set-object-1'));

    await waitFor(() => {
      expect(requests.length).toBeGreaterThan(0);
    });
    const firstRequestCount = requests.length;
    const firstSignals = requests.map((request) => request.signal);

    await fireEvent.click(screen.getByTestId('set-sps-2'));
    await fireEvent.click(screen.getByTestId('set-object-2'));

    await waitFor(() => {
      expect(requests.length).toBeGreaterThan(firstRequestCount);
    });
    expect(firstSignals.some((signal) => signal.aborted)).toBe(true);

    const latestRequestIndex = requests.length - 1;
    requests.forEach((request, index) => {
      request.resolve({ available: index === latestRequestIndex ? [22] : [11] });
    });

    await waitFor(() => {
      expect(screen.getByTestId('available-numbers')).toHaveTextContent('22');
    });
  });

  it('loads selected object-data preview', async () => {
    mockGetObjectData.mockResolvedValueOnce({ id: 'obj-1', bacnet_objects: [{ id: 'bo-1' }] });

    render(MultiCreateStateHarness);

    await fireEvent.click(screen.getByTestId('set-sps-1'));
    await fireEvent.click(screen.getByTestId('set-object-1'));

    await waitFor(() => {
      expect(mockGetObjectData).toHaveBeenCalledWith('obj-1', expect.any(AbortSignal));
      expect(screen.getByTestId('object-data-id')).toHaveTextContent('obj-1');
    });
  });
});
