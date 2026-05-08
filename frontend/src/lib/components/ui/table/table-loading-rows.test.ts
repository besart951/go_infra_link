/// <reference types="vitest" />

import { render } from '@testing-library/svelte';
import { tick } from 'svelte';
import TableLoadingRowsHarness from './__tests__/TableLoadingRowsHarness.svelte';

describe('Table.LoadingRows', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('waits before rendering loading rows', async () => {
    const rendered = render(TableLoadingRowsHarness, { loading: true, delayMs: 150 });

    await tick();
    expect(rendered.container.querySelectorAll('[data-table-loading-row]')).toHaveLength(0);

    await vi.advanceTimersByTimeAsync(149);
    await tick();
    expect(rendered.container.querySelectorAll('[data-table-loading-row]')).toHaveLength(0);

    await vi.advanceTimersByTimeAsync(1);
    await tick();
    expect(rendered.container.querySelectorAll('[data-table-loading-row]')).toHaveLength(2);
    expect(rendered.container.querySelectorAll('[data-table-loading-cell]')).toHaveLength(6);
  });

  it('cancels pending loading rows when loading finishes quickly', async () => {
    const rendered = render(TableLoadingRowsHarness, { loading: true, delayMs: 150 });

    await tick();
    await vi.advanceTimersByTimeAsync(100);
    await rendered.rerender({ loading: false, delayMs: 150 });
    await tick();

    await vi.advanceTimersByTimeAsync(100);
    await tick();
    expect(rendered.container.querySelectorAll('[data-table-loading-row]')).toHaveLength(0);
  });
});
