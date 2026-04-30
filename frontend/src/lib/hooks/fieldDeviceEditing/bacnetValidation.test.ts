import { describe, expect, it } from 'vitest';
import { buildBacnetObject, buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import { validateBacnetObjectEdits } from './bacnetValidation.js';
import type { BacnetObjectInput } from '$lib/domain/facility/index.js';

const translate = (key: string) => key;

function validate(edits: Partial<BacnetObjectInput>) {
  return validateBacnetObjectEdits({
    device: buildFieldDevice(),
    deviceEdits: new Map([['bo-1', edits]]),
    translate
  }).get('bo-1');
}

describe('field-device BACnet edit validation', () => {
  it.each([
    [
      'software_type',
      { software_type: 'bad-type' },
      'field_device.bacnet.validation.software_type'
    ],
    [
      'hardware_type',
      { hardware_type: 'bad-type' },
      'field_device.bacnet.validation.hardware_type'
    ],
    [
      'software_number',
      { software_number: -1 },
      'field_device.bacnet.validation.software_number_range'
    ],
    [
      'software_number',
      { software_number: 65536 },
      'field_device.bacnet.validation.software_number_range'
    ],
    [
      'hardware_quantity',
      { hardware_quantity: 0 },
      'field_device.bacnet.validation.hardware_quantity_range'
    ],
    [
      'hardware_quantity',
      { hardware_quantity: 256 },
      'field_device.bacnet.validation.hardware_quantity_range'
    ]
  ] as const)('rejects invalid %s edits', (field, edits, expectedError) => {
    expect(validate(edits)?.[field]).toBe(expectedError);
  });

  it('allows valid software and hardware type/range edits', () => {
    const errors = validateBacnetObjectEdits({
      device: buildFieldDevice(),
      deviceEdits: new Map([
        [
          'bo-1',
          {
            software_type: 'ao',
            software_number: 65535,
            hardware_type: 'do',
            hardware_quantity: 255
          }
        ]
      ]),
      translate
    });

    expect(errors.size).toBe(0);
  });

  it('flags duplicate effective software type and number combinations', () => {
    const errors = validateBacnetObjectEdits({
      device: buildFieldDevice({
        bacnet_objects: [
          buildBacnetObject({ id: 'bo-1', software_type: 'ai', software_number: 1 }),
          buildBacnetObject({ id: 'bo-2', software_type: 'ao', software_number: 2 })
        ]
      }),
      deviceEdits: new Map([['bo-2', { software_type: 'ai', software_number: 1 }]]),
      translate
    });

    expect(errors.get('bo-1')?.software_number).toBe(
      'field_device.bacnet.validation.software_unique'
    );
    expect(errors.get('bo-2')?.software_number).toBe(
      'field_device.bacnet.validation.software_unique'
    );
  });
});
