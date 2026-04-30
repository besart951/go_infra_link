/// <reference types="vitest" />

import { fireEvent, render, screen } from '@testing-library/svelte';
import FieldDeviceSearchBar from './FieldDeviceSearchBar.svelte';

const mockState = vi.hoisted(() => ({
  searchText: '',
  selectedCount: 0,
  showBulkEditPanel: false,
  showExportPanel: false,
  showFilterPanel: false,
  hasActiveFilters: false,
  loading: false,
  search: vi.fn(),
  clearSelection: vi.fn(),
  canDeleteFieldDevice: vi.fn(() => true),
  bulkDeleteSelected: vi.fn(),
  canOpenBulkEditPanel: vi.fn(() => true),
  toggleBulkEditPanel: vi.fn(),
  toggleExportPanel: vi.fn(),
  toggleFilterPanel: vi.fn(),
  reload: vi.fn()
}));

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string, params?: Record<string, unknown>) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('./state/context.svelte.js', () => ({
  useFieldDeviceState: () => mockState
}));

vi.mock('./FieldDeviceViewPopover.svelte', async () => ({
  default: (await import('./__tests__/mocks/Noop.svelte')).default
}));

describe('FieldDeviceSearchBar', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    mockState.searchText = '';
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('debounces field-device search by 300ms', async () => {
    render(FieldDeviceSearchBar);

    const input = screen.getByPlaceholderText('field_device.search.placeholder');
    await fireEvent.input(input, { target: { value: 'abc' } });

    expect(mockState.search).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(299);
    expect(mockState.search).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1);
    expect(mockState.search).toHaveBeenCalledTimes(1);
    expect(mockState.search).toHaveBeenCalledWith('abc');
  });

  it('uses the latest field-device search value during the debounce window', async () => {
    render(FieldDeviceSearchBar);

    const input = screen.getByPlaceholderText('field_device.search.placeholder');
    await fireEvent.input(input, { target: { value: 'a' } });
    await vi.advanceTimersByTimeAsync(150);
    await fireEvent.input(input, { target: { value: 'ab' } });
    await vi.advanceTimersByTimeAsync(150);
    await fireEvent.input(input, { target: { value: 'abc' } });

    await vi.advanceTimersByTimeAsync(300);
    expect(mockState.search).toHaveBeenCalledTimes(1);
    expect(mockState.search).toHaveBeenCalledWith('abc');
  });
});
