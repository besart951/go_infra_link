import { describe, expect, it } from 'vitest';
import {
  FieldDeviceCursorState,
  fieldDeviceCursorQueryKey
} from './FieldDeviceCursorState.svelte.js';
import type { DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from './types.js';

function query(filters: FieldDeviceFilters = {}): DataTableQuery<FieldDeviceFilters> {
  return { page: 1, pageSize: 300, searchText: '', filters };
}

describe('FieldDeviceCursorState', () => {
  it('moves in both directions and resets when the query changes', () => {
    const state = new FieldDeviceCursorState();
    expect(state.cursorFor(query())).toBeUndefined();
    state.apply({ items: [], nextCursor: 'next-1' });

    expect(state.moveNext()).toBe(true);
    expect(state.cursorFor(query())).toBe('next-1');
    state.apply({ items: [], nextCursor: 'next-2', previousCursor: 'previous-1' });
    expect(state.movePrevious()).toBe(true);
    expect(state.cursorFor(query())).toBe('previous-1');

    expect(state.cursorFor({ ...query(), searchText: 'pump' })).toBeUndefined();
    expect(state.hasNextPage).toBe(false);
    expect(state.hasPreviousPage).toBe(false);
  });

  it('uses a stable key for filters regardless of property order', () => {
    const first = fieldDeviceCursorQueryKey(query({ buildingId: 'b', projectId: 'p' }));
    const second = fieldDeviceCursorQueryKey(query({ projectId: 'p', buildingId: 'b' }));
    expect(first).toBe(second);
  });
});
