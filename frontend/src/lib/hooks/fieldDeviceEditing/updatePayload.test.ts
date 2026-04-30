import { describe, expect, it } from 'vitest';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import { buildFieldDeviceUpdatePayload } from './updatePayload.js';
import type { BacnetObjectInput, BulkUpdateFieldDeviceItem } from '$lib/domain/facility/index.js';

function buildPayload(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>> = new Map()
) {
  return buildFieldDeviceUpdatePayload({
    deviceId: 'fd-1',
    storeItems: [buildFieldDevice()],
    pendingEdits,
    pendingBacnetEdits,
    includeBacnet: true
  });
}

describe('field-device update payload construction', () => {
  it('returns null for an empty payload', () => {
    expect(buildPayload(new Map())).toBeNull();
  });

  it('builds base-only payloads', () => {
    expect(
      buildPayload(
        new Map([
          [
            'fd-1',
            {
              bmk: 'FD-UPDATED',
              description: 'Updated description',
              apparat_nr: 12
            }
          ]
        ])
      )
    ).toEqual({
      id: 'fd-1',
      bmk: 'FD-UPDATED',
      description: 'Updated description',
      apparat_nr: 12
    });
  });

  it('builds specification-only payloads', () => {
    expect(
      buildPayload(
        new Map([
          [
            'fd-1',
            {
              specification: {
                specification_brand: 'Brand X',
                electrical_connection_power: undefined
              }
            }
          ]
        ])
      )
    ).toEqual({
      id: 'fd-1',
      specification: {
        specification_brand: 'Brand X'
      }
    });
  });

  it('builds mixed base, specification, and BACnet payloads', () => {
    expect(
      buildPayload(
        new Map([
          [
            'fd-1',
            {
              text_fix: 'TXT-UPDATED',
              specification: {
                specification_supplier: 'Supplier X'
              }
            }
          ]
        ]),
        new Map([['fd-1', new Map([['bo-1', { software_number: 42 }]])]])
      )
    ).toEqual({
      id: 'fd-1',
      text_fix: 'TXT-UPDATED',
      specification: {
        specification_supplier: 'Supplier X'
      },
      bacnet_objects: [{ id: 'bo-1', software_number: 42 }]
    });
  });
});
