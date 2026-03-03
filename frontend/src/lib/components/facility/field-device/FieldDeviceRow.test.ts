/// <reference types="vitest" />

import { render, screen } from '@testing-library/svelte';
import FieldDeviceRow from './FieldDeviceRow.svelte';

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

describe('FieldDeviceRow', () => {
  it('renders empty form values for a new row', () => {
    render(FieldDeviceRow, {
      index: 0,
      row: { id: 'r-1', bmk: '', description: '', textFix: '', apparatNr: null },
      error: null,
      placeholder: '',
      onBmkChange: vi.fn(),
      onDescriptionChange: vi.fn(),
      onTextFixChange: vi.fn(),
      onApparatNrChange: vi.fn(),
      onRemove: vi.fn()
    });

    expect(screen.getByLabelText('field_device.row.bmk')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.description')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.text_fix')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.apparat_nr')).toHaveValue(null);
  });

  it('populates form fields correctly for existing row data', () => {
    render(FieldDeviceRow, {
      index: 0,
      row: {
        id: 'r-2',
        bmk: 'BMK-EXISTING',
        description: 'Existing device description',
        textFix: 'TF-EX',
        apparatNr: 12
      },
      error: null,
      placeholder: '',
      onBmkChange: vi.fn(),
      onDescriptionChange: vi.fn(),
      onTextFixChange: vi.fn(),
      onApparatNrChange: vi.fn(),
      onRemove: vi.fn()
    });

    expect(screen.getByLabelText('field_device.row.bmk')).toHaveValue('BMK-EXISTING');
    expect(screen.getByLabelText('field_device.row.description')).toHaveValue(
      'Existing device description'
    );
    expect(screen.getByLabelText('field_device.row.text_fix')).toHaveValue('TF-EX');
    expect(screen.getByLabelText('field_device.row.apparat_nr')).toHaveValue(12);
  });
});
