import { describe, expect, it, beforeEach, vi } from 'vitest';
import {
  decodeMultiFilter,
  encodeMultiFilter,
  ProjectFacilityListFilterStore,
  sanitizeMultiFilterValue
} from './projectFacilityListFilters.js';

function installLocalStorageMock(): void {
  const data = new Map<string, string>();
  const storage = {
    get length() {
      return data.size;
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => [...data.keys()][index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value))
  } satisfies Storage;

  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: storage
  });
  Object.defineProperty(window, 'localStorage', {
    configurable: true,
    value: storage
  });
}

describe('project facility list filters', () => {
  beforeEach(() => {
    installLocalStorageMock();
  });

  it('encodes multi-select values without duplicates or blank entries', () => {
    const encoded = encodeMultiFilter(['building-1', ' ', 'building-1', 'building/2']);

    expect(encoded).toBe('building-1|building%2F2');
    expect(decodeMultiFilter(encoded)).toEqual(['building-1', 'building/2']);
  });

  it('drops stored selections that are no longer available', () => {
    expect(
      sanitizeMultiFilterValue(
        encodeMultiFilter(['building-1', 'building-2']),
        new Set(['building-2'])
      )
    ).toBe('building-2');
  });

  it('stores filters per project and removes empty selections', () => {
    const store = new ProjectFacilityListFilterStore<{ buildingIds?: string }>('test-scope');

    store.save('project-1', { buildingIds: 'building-1' });
    store.save('project-2', { buildingIds: 'building-2' });

    expect(store.load('project-1')).toEqual({ buildingIds: 'building-1' });
    expect(store.load('project-2')).toEqual({ buildingIds: 'building-2' });

    store.save('project-1', {});

    expect(store.load('project-1')).toEqual({});
    expect(store.load('project-2')).toEqual({ buildingIds: 'building-2' });
  });
});
