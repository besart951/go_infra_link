import { describe, expect, it } from 'vitest';
import {
  buildSpecificationPatch,
  normalizeSpecificationForDisplay,
  toDisplayOptionalValue
} from './specificationEdits.js';

describe('field-device specification edits', () => {
  it('builds no patch for missing or only-undefined specification edits', () => {
    expect(buildSpecificationPatch(undefined)).toBeUndefined();
    expect(
      buildSpecificationPatch({
        specification_brand: undefined,
        electrical_connection_power: undefined
      })
    ).toBeUndefined();
  });

  it('keeps null and defined values while dropping undefined values', () => {
    expect(
      buildSpecificationPatch({
        specification_brand: 'Brand X',
        specification_supplier: null,
        electrical_connection_power: undefined
      })
    ).toEqual({
      specification_brand: 'Brand X',
      specification_supplier: null
    });
  });

  it('normalizes null optional values for display', () => {
    expect(toDisplayOptionalValue(null)).toBeUndefined();
    expect(toDisplayOptionalValue('Brand X')).toBe('Brand X');
    expect(
      normalizeSpecificationForDisplay({
        specification_brand: null,
        specification_supplier: 'Supplier X'
      })
    ).toEqual({
      specification_brand: undefined,
      specification_supplier: 'Supplier X'
    });
  });
});
