import { describe, expect, it } from 'vitest';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import { reconcileFieldDeviceSaveResult } from './saveReconciliation.js';
import type { BacnetObjectInput, BulkUpdateFieldDeviceItem } from '$lib/domain/facility/index.js';

const localizeEditErrorInfo = <T>(info?: T) => info;
const localizeFieldErrorMap = (fields: Record<string, string>) => fields;

describe('field-device save-result reconciliation', () => {
  it('clears fully successful draft state', () => {
    const pendingEdits = new Map<string, Partial<BulkUpdateFieldDeviceItem>>([
      ['fd-1', { bmk: 'FD-SAVED' }]
    ]);
    const pendingBacnetEdits = new Map<string, Map<string, Partial<BacnetObjectInput>>>([
      ['fd-1', new Map([['bo-1', { software_number: 42 }]])]
    ]);

    const result = reconcileFieldDeviceSaveResult({
      storeItems: [buildFieldDevice()],
      updates: [
        { id: 'fd-1', bmk: 'FD-SAVED', bacnet_objects: [{ id: 'bo-1', software_number: 42 }] }
      ],
      result: {
        results: [{ id: 'fd-1', success: true }],
        total_count: 1,
        success_count: 1,
        failure_count: 0
      },
      pendingEdits,
      pendingBacnetEdits,
      pendingEditsSnapshot: new Map(pendingEdits),
      pendingBacnetEditsSnapshot: new Map(pendingBacnetEdits),
      existingErrors: new Map([['fd-1', { message: 'old error' }]]),
      localizeEditErrorInfo,
      localizeFieldErrorMap
    });

    expect(result.remainingEdits.size).toBe(0);
    expect(result.remainingBacnetEdits.size).toBe(0);
    expect(result.editErrors.size).toBe(0);
    expect(result.successIds).toEqual(new Set(['fd-1']));
  });

  it('keeps failed phases while clearing successful phases after partial success', () => {
    const device = buildFieldDevice();
    const pendingEdits = new Map<string, Partial<BulkUpdateFieldDeviceItem>>([
      [
        'fd-1',
        {
          bmk: 'FD-PARTIAL',
          specification: {
            specification_brand: 'Rejected Brand'
          }
        }
      ]
    ]);
    const pendingBacnetEdits = new Map<string, Map<string, Partial<BacnetObjectInput>>>([
      ['fd-1', new Map([['bo-1', { text_fix: 'TF-PARTIAL' }]])]
    ]);

    const result = reconcileFieldDeviceSaveResult({
      storeItems: [device],
      updates: [
        {
          id: 'fd-1',
          bmk: 'FD-PARTIAL',
          specification: { specification_brand: 'Rejected Brand' },
          bacnet_objects: [{ id: 'bo-1', text_fix: 'TF-PARTIAL' }]
        }
      ],
      result: {
        results: [
          {
            id: 'fd-1',
            success: false,
            error: 'validation failed',
            fields: {
              'specification.specification_brand': 'brand rejected'
            }
          }
        ],
        total_count: 1,
        success_count: 0,
        failure_count: 1
      },
      pendingEdits,
      pendingBacnetEdits,
      pendingEditsSnapshot: new Map(pendingEdits),
      pendingBacnetEditsSnapshot: new Map(pendingBacnetEdits),
      existingErrors: new Map(),
      localizeEditErrorInfo,
      localizeFieldErrorMap
    });

    expect(result.remainingEdits.get('fd-1')).toEqual({
      specification: { specification_brand: 'Rejected Brand' }
    });
    expect(result.remainingBacnetEdits.size).toBe(0);
    expect(result.partialSuccessIds).toEqual(new Set(['fd-1']));
    expect(result.editErrors.get('fd-1')).toEqual({
      message: 'validation failed',
      fields: { 'specification.specification_brand': 'brand rejected' }
    });
    expect(result.optimisticUpdates[0].bmk).toBe('FD-PARTIAL');
    expect(result.optimisticUpdates[0].specification?.specification_brand).toBe(
      device.specification?.specification_brand
    );
    expect(result.optimisticUpdates[0].bacnet_objects?.[0].text_fix).toBe('TF-PARTIAL');
  });
});
