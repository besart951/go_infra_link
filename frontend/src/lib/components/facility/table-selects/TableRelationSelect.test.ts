import { fireEvent, render, screen } from '@testing-library/svelte';
import TableApparatSelect from './TableApparatSelect.svelte';
import TableSystemPartSelect from './TableSystemPartSelect.svelte';

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

describe('facility table relation selects', () => {
  it('shows the selected apparat as short_name - name', () => {
    render(TableApparatSelect, {
      items: [
        {
          id: 'apparat-1',
          short_name: 'Ven',
          name: 'Ventilator',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ],
      value: 'apparat-1'
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('Ven - Ventilator');
  });

  it('shows the selected system part as short_name - name', () => {
    render(TableSystemPartSelect, {
      items: [
        {
          id: 'system-part-1',
          short_name: 'Ele',
          name: 'Elektro',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ],
      value: 'system-part-1'
    });

    expect(screen.getByRole('combobox')).toHaveTextContent('Ele - Elektro');
  });

  it('can clear the selected apparat when enabled', async () => {
    const onValueChange = vi.fn();
    render(TableApparatSelect, {
      items: [
        {
          id: 'apparat-1',
          short_name: 'Ven',
          name: 'Ventilator',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ],
      value: 'apparat-1',
      clearable: true,
      onValueChange
    });

    await fireEvent.click(screen.getByRole('combobox'));
    await fireEvent.click(await screen.findByText('field_device.table_select.clear_apparat'));

    expect(onValueChange).toHaveBeenCalledWith('');
  });
});
