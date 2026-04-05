/// <reference types="vitest" />

import { render, screen } from '@testing-library/svelte';
import FieldDeviceRow from './FieldDeviceRow.svelte';

import type { FieldDeviceRowData } from '$lib/domain/facility/fieldDeviceMultiCreate.js';
import type { FieldDeviceMultiCreateState } from './multi-create/FieldDeviceMultiCreateState.svelte.js';

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(fn: (value: (key: string) => string) => void) {
      fn((key: string) => key);
      return () => {};
    }
  })
}));

describe('FieldDeviceRow', () => {
  function createMockState(rows: FieldDeviceRowData[]): FieldDeviceMultiCreateState {
    const mock = {
      rows,
      rowErrors: new Map(),
      submitting: false,
      getPlaceholderForRow: function (): string {
        return '';
      },
      handleRowBmkChange: vi.fn(),
      handleRowDescriptionChange: vi.fn(),
      handleRowTextFixChange: vi.fn(),
      handleRowApparatNrChange: vi.fn(),
      removeRow: vi.fn()
    };

    return mock as unknown as FieldDeviceMultiCreateState;
  }

  it('renders empty form values for a new row', () => {
    const state = createMockState([
      { id: 'r-1', bmk: '', description: '', textFix: '', apparatNr: null }
    ]);

    render(FieldDeviceRow, {
      index: 0,
      state
    });

    expect(screen.getByLabelText('field_device.row.bmk')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.description')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.text_fix')).toHaveValue('');
    expect(screen.getByLabelText('field_device.row.apparat_nr')).toHaveValue(null);
  });

  it('populates form fields correctly for existing row data', () => {
    const state = createMockState([
      {
        id: 'r-2',
        bmk: 'BMK-EXISTING',
        description: 'Existing device description',
        textFix: 'TF-EX',
        apparatNr: 12
      }
    ]);

    render(FieldDeviceRow, {
      index: 0,
      state
    });

    expect(screen.getByLabelText('field_device.row.bmk')).toHaveValue('BMK-EXISTING');
    expect(screen.getByLabelText('field_device.row.description')).toHaveValue(
      'Existing device description'
    );
    expect(screen.getByLabelText('field_device.row.text_fix')).toHaveValue('TF-EX');
    expect(screen.getByLabelText('field_device.row.apparat_nr')).toHaveValue(12);
  });
});
