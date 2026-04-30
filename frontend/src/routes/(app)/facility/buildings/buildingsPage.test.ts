/// <reference types="vitest" />

import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import BuildingsPage from './+page.svelte';
import type { Building } from '$lib/domain/facility/index.js';

const mockListStore = vi.hoisted(() => {
  const state = {
    items: [] as Building[],
    loading: false,
    error: null as string | null,
    page: 1,
    totalPages: 1,
    total: 0,
    searchText: ''
  };
  const subscribers = new Set<(value: typeof state) => void>();

  return {
    state,
    subscribers,
    load: vi.fn(),
    reload: vi.fn(),
    search: vi.fn(),
    goToPage: vi.fn(),
    setItems(items: Building[]) {
      state.items = items;
      state.total = items.length;
      state.totalPages = 1;
      for (const subscriber of subscribers) subscriber(state);
    }
  };
});

const mockRuntime = vi.hoisted(() => ({
  deleteBuilding: vi.fn(),
  confirm: vi.fn(),
  addToast: vi.fn(),
  goto: vi.fn(),
  grants: new Set<string>()
}));

vi.mock('$app/navigation', () => ({
  goto: (...args: unknown[]) => mockRuntime.goto(...args)
}));

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/components/confirm-dialog.svelte', async () => ({
  default: (await import('./__tests__/Noop.svelte')).default
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: (...args: unknown[]) => mockRuntime.addToast(...args)
}));

vi.mock('$lib/stores/confirm-dialog.js', () => ({
  confirm: (...args: unknown[]) => mockRuntime.confirm(...args)
}));

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => mockRuntime.grants.has(`${resource}.${action}`)
}));

vi.mock('$lib/stores/list/entityStores.js', () => ({
  buildingsStore: {
    subscribe(fn: (value: typeof mockListStore.state) => void) {
      fn(mockListStore.state);
      mockListStore.subscribers.add(fn);
      return () => mockListStore.subscribers.delete(fn);
    },
    load: (...args: unknown[]) => mockListStore.load(...args),
    reload: (...args: unknown[]) => mockListStore.reload(...args),
    search: (...args: unknown[]) => mockListStore.search(...args),
    goToPage: (...args: unknown[]) => mockListStore.goToPage(...args)
  }
}));

vi.mock('$lib/infrastructure/api/buildingRepository.js', () => ({
  buildingRepository: {
    delete: (...args: unknown[]) => mockRuntime.deleteBuilding(...args),
    create: vi.fn(),
    update: vi.fn(),
    get: vi.fn(),
    list: vi.fn(),
    validate: vi.fn()
  }
}));

vi.mock('$lib/components/facility/forms/BuildingForm.svelte', async () => ({
  default: (await import('./__tests__/MockBuildingForm.svelte')).default
}));

function grant(...permissions: string[]) {
  mockRuntime.grants = new Set(permissions);
}

function building(overrides: Partial<Building> = {}): Building {
  return {
    id: 'building-1',
    iws_code: 'IWS-100',
    building_group: 7,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  };
}

describe('buildings facility page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockListStore.subscribers.clear();
    mockListStore.setItems([building()]);
    mockRuntime.confirm.mockResolvedValue(true);
    mockRuntime.deleteBuilding.mockResolvedValue(undefined);
    grant('building.create', 'building.update', 'building.delete');
  });

  it('opens create form and reloads after form success', async () => {
    render(BuildingsPage);

    await waitFor(() => expect(mockListStore.load).toHaveBeenCalledTimes(1));
    await fireEvent.click(screen.getByRole('button', { name: /facility.new_building/ }));

    expect(screen.getByTestId('form-mode')).toHaveTextContent('create');

    await fireEvent.click(screen.getByText('form-success'));

    await waitFor(() => expect(mockListStore.reload).toHaveBeenCalledTimes(1));
    expect(screen.queryByTestId('building-form')).not.toBeInTheDocument();
  });

  it('opens edit form from row action menu', async () => {
    const rendered = render(BuildingsPage);

    await fireEvent.click(rendered.container.querySelector('tbody button')!);
    await fireEvent.click(await screen.findByText('common.edit'));

    expect(screen.getByTestId('form-mode')).toHaveTextContent('building-1');
  });

  it('deletes row after confirmation and reloads list', async () => {
    const rendered = render(BuildingsPage);

    await fireEvent.click(rendered.container.querySelector('tbody button')!);
    await fireEvent.click(await screen.findByText('common.delete'));

    await waitFor(() => {
      expect(mockRuntime.confirm).toHaveBeenCalledWith(
        expect.objectContaining({
          message: 'facility.delete_building_confirm'
        })
      );
      expect(mockRuntime.deleteBuilding).toHaveBeenCalledWith('building-1', undefined);
      expect(mockRuntime.addToast).toHaveBeenCalledWith('facility.building_deleted', 'success');
      expect(mockListStore.reload).toHaveBeenCalledTimes(1);
    });
  });

  it('hides create, edit, and delete when permission is missing', async () => {
    grant();
    const rendered = render(BuildingsPage);

    expect(screen.queryByRole('button', { name: /facility.new_building/ })).not.toBeInTheDocument();

    await fireEvent.click(rendered.container.querySelector('tbody button')!);

    expect(screen.queryByText('common.edit')).not.toBeInTheDocument();
    expect(screen.queryByText('common.delete')).not.toBeInTheDocument();
  });
});
