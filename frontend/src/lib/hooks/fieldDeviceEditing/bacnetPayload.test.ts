import { describe, expect, it } from 'vitest';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import { buildBacnetObjectsPayload } from './bacnetPayload.js';

describe('field-device BACnet payload construction', () => {
  it('builds a full create-shaped object patch when every editable field changed', () => {
    const device = buildFieldDevice();

    expect(
      buildBacnetObjectsPayload(
        device,
        new Map([
          [
            'bo-1',
            {
              text_fix: 'TF-100',
              description: 'Description',
              gms_visible: false,
              optional: true,
              text_individual: 'Individual',
              software_type: 'ao',
              software_number: 100,
              hardware_type: 'do',
              hardware_quantity: 3,
              software_reference_id: 'bo-ref',
              state_text_id: 'state-1',
              notification_class_id: 'notification-1',
              alarm_type_id: 'alarm-1'
            }
          ]
        ])
      )
    ).toEqual([
      {
        id: 'bo-1',
        text_fix: 'TF-100',
        description: 'Description',
        gms_visible: false,
        optional: true,
        text_individual: 'Individual',
        software_type: 'ao',
        software_number: 100,
        hardware_type: 'do',
        hardware_quantity: 3,
        software_reference_id: 'bo-ref',
        state_text_id: 'state-1',
        notification_class_id: 'notification-1',
        alarm_type_id: 'alarm-1'
      }
    ]);
  });

  it('builds a partial update-shaped object patch', () => {
    expect(
      buildBacnetObjectsPayload(
        buildFieldDevice(),
        new Map([['bo-1', { software_number: 42, alarm_type_id: undefined }]])
      )
    ).toEqual([{ id: 'bo-1', software_number: 42, alarm_type_id: undefined }]);
  });

  it('ignores delete-shaped id-only patches because bulk field-device patch does not delete BACnet objects', () => {
    expect(buildBacnetObjectsPayload(buildFieldDevice(), new Map([['bo-1', {}]]))).toEqual([]);
  });
});
