import { describe, expect, it } from 'vitest';
import { getFailedPhases, getUpdatePhases, hasPartialPhaseSuccess } from './bulkUpdatePhases.js';

describe('field-device bulk update phase classification', () => {
  it('detects base, specification, and BACnet update phases', () => {
    expect(
      getUpdatePhases({
        id: 'fd-1',
        bmk: 'FD-1',
        specification: { specification_brand: 'Brand X' },
        bacnet_objects: [{ id: 'bo-1', software_number: 42 }]
      })
    ).toEqual(new Set(['fielddevice', 'specification', 'bacnet_objects']));
  });

  it('detects failed phases from field paths', () => {
    expect(
      getFailedPhases({
        'fielddevice.apparat_nr': 'invalid',
        'specification.specification_brand': 'invalid',
        'bacnet_objects.bo-1.software_number': 'invalid'
      })
    ).toEqual(new Set(['fielddevice', 'specification', 'bacnet_objects']));
  });

  it('reports partial success when at least one update phase has no failure', () => {
    expect(
      hasPartialPhaseSuccess(
        {
          id: 'fd-1',
          bmk: 'FD-1',
          specification: { specification_brand: 'Brand X' },
          bacnet_objects: [{ id: 'bo-1', software_number: 42 }]
        },
        {
          'specification.specification_brand': 'invalid'
        }
      )
    ).toBe(true);
  });

  it('reports no partial success when all update phases failed', () => {
    expect(
      hasPartialPhaseSuccess(
        {
          id: 'fd-1',
          bmk: 'FD-1',
          specification: { specification_brand: 'Brand X' }
        },
        {
          fielddevice: 'invalid',
          specification: 'invalid'
        }
      )
    ).toBe(false);
  });
});
