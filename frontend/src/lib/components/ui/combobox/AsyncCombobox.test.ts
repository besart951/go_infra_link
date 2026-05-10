import { render, screen, waitFor } from '@testing-library/svelte';
import AsyncCombobox from './AsyncCombobox.svelte';

const Combobox = AsyncCombobox as any;

type Item = {
  id: string;
  label: string;
};

const items: Item[] = [{ id: 'app-1', label: 'AHU - Air Handling Unit' }];

describe('AsyncCombobox', () => {
  beforeAll(() => {
    HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  it('does not expose the raw selected id while the selected label is loading', async () => {
    let resolveSelected!: (item: Item) => void;
    const fetchById = vi.fn(
      () =>
        new Promise<Item>((resolve) => {
          resolveSelected = resolve;
        })
    );

    render(Combobox, {
      value: 'app-1',
      fetcher: vi.fn().mockResolvedValue(items),
      fetchById,
      labelKey: 'label',
      placeholder: 'Choose item'
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('Choose item');
    expect(screen.getByRole('combobox')).not.toHaveTextContent('app-1');

    resolveSelected(items[0]);

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toHaveTextContent('AHU - Air Handling Unit');
    });
  });

  it('keeps the selected label visible when refreshKey changes', async () => {
    const rendered = render(Combobox, {
      value: 'app-1',
      fetcher: vi.fn().mockResolvedValue(items),
      fetchById: vi.fn().mockResolvedValue(items[0]),
      labelKey: 'label',
      refreshKey: 'initial',
      placeholder: 'Choose item'
    });

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toHaveTextContent('AHU - Air Handling Unit');
    });

    await rendered.rerender({
      value: 'app-1',
      fetcher: vi.fn().mockResolvedValue(items),
      fetchById: vi.fn().mockResolvedValue(items[0]),
      labelKey: 'label',
      refreshKey: 'changed',
      placeholder: 'Choose item'
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('AHU - Air Handling Unit');
    expect(screen.getByRole('combobox')).not.toHaveTextContent('app-1');
  });
});
