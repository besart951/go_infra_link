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

  it('keeps the complete device draft when an atomic item fails', () => {
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
      bmk: 'FD-PARTIAL',
      specification: { specification_brand: 'Rejected Brand' }
    });
    expect(result.remainingBacnetEdits.get('fd-1')?.get('bo-1')).toEqual({
      text_fix: 'TF-PARTIAL'
    });
    expect(result.successIds.size).toBe(0);
    expect(result.editErrors.get('fd-1')).toEqual({
      message: 'validation failed',
      fields: { 'specification.specification_brand': 'brand rejected' }
    });
    expect(result.optimisticUpdates).toEqual([]);
  });

  it('understands indexed update paths from unified validation errors', () => {
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
      ['fd-1', new Map([['bo-1', { alarm_type_id: 'alarm-9' }]])]
    ]);

    const result = reconcileFieldDeviceSaveResult({
      storeItems: [device],
      updates: [
        {
          id: 'fd-1',
          bmk: 'FD-PARTIAL',
          specification: { specification_brand: 'Rejected Brand' },
          bacnet_objects: [{ id: 'bo-1', alarm_type_id: 'alarm-9' }]
        }
      ],
      result: {
        results: [
          {
            id: 'fd-1',
            success: false,
            error: 'validation failed',
            fields: {
              'updates[0].specifications.specification_brand': 'brand rejected',
              'updates[0].bacnet_objects[0].alarm_type_id': 'alarm type rejected'
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
      bmk: 'FD-PARTIAL',
      specification: { specification_brand: 'Rejected Brand' }
    });
    expect(result.remainingBacnetEdits.get('fd-1')?.get('bo-1')).toEqual({
      alarm_type_id: 'alarm-9'
    });
    expect(result.bacnetFieldErrors.get('fd-1')?.get('bo-1')?.alarm_type_id).toBe(
      'alarm type rejected'
    );
  });
});
