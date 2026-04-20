/// <reference types="vitest" />

import { render } from '@testing-library/svelte';
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

    const { container } = render(FieldDeviceRow, {
      index: 0,
      state
    });

    expect(container.querySelector('#bmk-0')).toHaveValue('');
    expect(container.querySelector('#description-0')).toHaveValue('');
    expect(container.querySelector('#text-fix-0')).toHaveValue('');
    expect(container.querySelector('#apparat-nr-0')).toHaveValue(null);
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

    const { container } = render(FieldDeviceRow, {
      index: 0,
      state
    });

    expect(container.querySelector('#bmk-0')).toHaveValue('BMK-EXISTING');
    expect(container.querySelector('#description-0')).toHaveValue('Existing device description');
    expect(container.querySelector('#text-fix-0')).toHaveValue('TF-EX');
    expect(container.querySelector('#apparat-nr-0')).toHaveValue(12);
  });
});
