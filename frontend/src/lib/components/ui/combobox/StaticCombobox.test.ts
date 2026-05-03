import { fireEvent, render, screen } from '@testing-library/svelte';
import StaticCombobox from './StaticCombobox.svelte';

const Combobox = StaticCombobox as any;

const items = [
  { id: 'app-1', label: 'Apparat 1' },
  { id: 'app-2', label: 'Apparat 2' }
];

describe('StaticCombobox', () => {
  beforeAll(() => {
    HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  it('reflects parent value resets after a local selection', async () => {
    const onValueChange = vi.fn();
    const rendered = render(Combobox, {
      items,
      value: 'app-1',
      labelKey: 'label',
      width: 'w-full',
      onValueChange
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('Apparat 1');

    await fireEvent.click(screen.getByRole('combobox'));
    await fireEvent.click(await screen.findByText('Apparat 2'));

    expect(onValueChange).toHaveBeenCalledWith('app-2');
    expect(screen.getByRole('combobox')).toHaveTextContent('Apparat 2');

    await rendered.rerender({
      items,
      value: 'app-1',
      labelKey: 'label',
      width: 'w-full',
      onValueChange
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('Apparat 1');
  });
});
