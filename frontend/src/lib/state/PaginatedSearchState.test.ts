import { describe, expect, it } from 'vitest';
import { PaginatedSearchState } from './PaginatedSearchState.svelte.js';

describe('PaginatedSearchState', () => {
  it('starts with consistent pagination and search defaults', () => {
    const state = new PaginatedSearchState({ pageSize: 25, initialLoading: true });

    expect(state.snapshot()).toEqual({
      total: 0,
      page: 1,
      pageSize: 25,
      totalPages: 0,
      searchText: '',
      loading: true,
      error: null
    });
  });

  it('resets to the first page when search text changes', () => {
    const state = new PaginatedSearchState();
    state.goToPage(4);

    state.applySearch('ada');

    expect(state.page).toBe(1);
    expect(state.searchText).toBe('ada');
  });

  it('applies paginated result metadata without changing search unless provided', () => {
    const state = new PaginatedSearchState();
    state.applySearch('ada');

    state.applyResult({ total: 42, page: 2, totalPages: 5 });

    expect(state.total).toBe(42);
    expect(state.page).toBe(2);
    expect(state.totalPages).toBe(5);
    expect(state.searchText).toBe('ada');
  });

  it('normalizes invalid pages to the first page', () => {
    const state = new PaginatedSearchState();

    state.goToPage(-10);

    expect(state.page).toBe(1);
  });
});
