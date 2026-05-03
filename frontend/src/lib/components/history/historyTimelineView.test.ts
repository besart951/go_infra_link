import { describe, expect, it } from 'vitest';
import type { ChangeEvent } from '$lib/domain/history.js';
import { buildHistoryTimelineView } from './historyTimelineView.js';

const cabinetID = 'cabinet-1';
const spsID = 'sps-1';
const fieldDeviceID = 'field-device-1';
const bacnetObjectID = 'bacnet-object-1';

function event(overrides: Partial<ChangeEvent>): ChangeEvent {
  return {
    id: crypto.randomUUID(),
    occurred_at: '2026-05-03T09:00:00Z',
    action: 'update',
    entity_table: 'control_cabinets',
    entity_id: cabinetID,
    diff_json: {
      control_cabinet_nr: {
        before: 'SGK33',
        after: 'SGK99'
      }
    },
    ...overrides
  };
}

describe('buildHistoryTimelineView', () => {
  it('keeps parent building events out of the control cabinet timeline', () => {
    const view = buildHistoryTimelineView(
      [
        event({
          entity_table: 'buildings',
          entity_id: 'building-1',
          diff_json: { name: { before: 'A', after: 'B' } }
        }),
        event({ id: 'cabinet-event' })
      ],
      { scopeType: 'control_cabinet', scopeId: cabinetID, controlCabinetId: cabinetID }
    );

    expect(view.directRows.map((row) => row.event.id)).toEqual(['cabinet-event']);
    expect(view.childGroups).toHaveLength(0);
  });

  it('groups child changes below their SPS controller scope', () => {
    const view = buildHistoryTimelineView(
      [
        event({
          id: 'sps-event',
          entity_table: 'sps_controllers',
          entity_id: spsID,
          diff_json: { device_name: { before: 'AS01', after: 'AS02' } },
          scopes: [{ scope_type: 'sps_controller', scope_id: spsID, label: 'AS02' }]
        }),
        event({
          id: 'field-device-event',
          entity_table: 'field_devices',
          entity_id: fieldDeviceID,
          diff_json: { bmk: { before: 'A', after: 'B' } },
          scopes: [
            { scope_type: 'sps_controller', scope_id: spsID, label: 'AS02' },
            { scope_type: 'field_device', scope_id: fieldDeviceID, label: 'B01' }
          ]
        })
      ],
      { scopeType: 'control_cabinet', scopeId: cabinetID, controlCabinetId: cabinetID }
    );

    expect(view.directRows).toHaveLength(0);
    expect(view.childGroups).toHaveLength(1);
    expect(view.childGroups[0].key).toBe(`sps_controller:${spsID}`);
    expect(view.childGroups[0].label).toBe('AS02');
    expect(view.childGroups[0].rows.map((row) => row.event.id)).toEqual(['sps-event']);
    expect(view.childGroups[0].children[0].key).toBe(`field_device:${fieldDeviceID}`);
    expect(view.childGroups[0].children[0].rows.map((row) => row.event.id)).toEqual([
      'field-device-event'
    ]);
  });

  it('keeps field device and specification changes together while nesting BACnet objects', () => {
    const view = buildHistoryTimelineView(
      [
        event({
          id: 'field-device-event',
          entity_table: 'field_devices',
          entity_id: fieldDeviceID,
          diff_json: { text_fix: { before: 'ALT', after: 'NEU' } },
          scopes: [{ scope_type: 'field_device', scope_id: fieldDeviceID, label: 'B01' }]
        }),
        event({
          id: 'specification-event',
          entity_table: 'specifications',
          entity_id: 'specification-1',
          diff_json: { specification_type: { before: 'A', after: 'B' } },
          scopes: [{ scope_type: 'field_device', scope_id: fieldDeviceID, label: 'B01' }]
        }),
        event({
          id: 'bacnet-event',
          entity_table: 'bacnet_objects',
          entity_id: bacnetObjectID,
          diff_json: { text_fix: { before: 'AI1', after: 'AI2' } },
          scopes: [
            { scope_type: 'field_device', scope_id: fieldDeviceID, label: 'B01' },
            { scope_type: 'bacnet_object', scope_id: bacnetObjectID, label: 'AI2' }
          ]
        })
      ],
      { scopeType: 'field_device', scopeId: fieldDeviceID }
    );

    expect(view.directRows.map((row) => row.event.id)).toEqual([
      'field-device-event',
      'specification-event'
    ]);
    expect(view.childGroups).toHaveLength(1);
    expect(view.childGroups[0].key).toBe(`bacnet_object:${bacnetObjectID}`);
    expect(view.childGroups[0].rows.map((row) => row.event.id)).toEqual(['bacnet-event']);
  });
});
