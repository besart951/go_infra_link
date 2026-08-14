import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import AsyncMultiSelect from './AsyncMultiSelect.svelte';

const MultiSelect = AsyncMultiSelect as any;

type Item = { id: string; label: string };

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((promiseResolve) => {
    resolve = promiseResolve;
  });

  return { promise, resolve };
}

describe('AsyncMultiSelect', () => {
  it('keeps refreshed options when an earlier request resolves last', async () => {
    const first = deferred<Item[]>();
    const second = deferred<Item[]>();
    const fetcher = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const rendered = render(MultiSelect, {
      value: [],
      fetcher,
      labelKey: 'label',
      placeholder: 'Choose items',
      refreshKey: 0
    });

    await fireEvent.click(screen.getByRole('combobox'));
    await waitFor(() => {
      expect(fetcher).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      value: [],
      fetcher,
      labelKey: 'label',
      placeholder: 'Choose items',
      refreshKey: 1
    });

    await waitFor(() => {
      expect(fetcher).toHaveBeenCalledTimes(2);
    });

    second.resolve([{ id: 'apparat-2', label: 'Fresh option' }]);
    await screen.findByText('Fresh option');

    first.resolve([{ id: 'apparat-1', label: 'Stale option' }]);
    await waitFor(() => {
      expect(screen.getByText('Fresh option')).toBeInTheDocument();
      expect(screen.queryByText('Stale option')).not.toBeInTheDocument();
    });
  });

  it('keeps refreshed selected items when an earlier request resolves last', async () => {
    const first = deferred<Item[]>();
    const second = deferred<Item[]>();
    const fetchByIds = vi
      .fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const selectedIDs = ['apparat-1'];
    const rendered = render(MultiSelect, {
      value: selectedIDs,
      fetcher: vi.fn().mockResolvedValue([]),
      fetchByIds,
      labelKey: 'label',
      placeholder: 'Choose items',
      refreshKey: 0
    });

    await waitFor(() => {
      expect(fetchByIds).toHaveBeenCalledTimes(1);
    });

    await rendered.rerender({
      value: selectedIDs,
      fetcher: vi.fn().mockResolvedValue([]),
      fetchByIds,
      labelKey: 'label',
      placeholder: 'Choose items',
      refreshKey: 1
    });

    await waitFor(() => {
      expect(fetchByIds).toHaveBeenCalledTimes(2);
    });

    second.resolve([{ id: 'apparat-1', label: 'Fresh selection' }]);
    await screen.findByText('Fresh selection');

    first.resolve([{ id: 'apparat-1', label: 'Stale selection' }]);
    await waitFor(() => {
      expect(screen.getByText('Fresh selection')).toBeInTheDocument();
      expect(screen.queryByText('Stale selection')).not.toBeInTheDocument();
    });
  });
});
