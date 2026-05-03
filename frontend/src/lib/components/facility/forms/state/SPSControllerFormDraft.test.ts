import { describe, expect, it } from 'vitest';
import type { Building, ControlCabinet, SystemType } from '$lib/domain/facility/index.js';
import {
  buildSPSControllerDeviceName,
  buildSPSControllerSystemTypeLabel,
  getNextAvailableSystemTypeNumber,
  getSPSControllerSystemTypeAddState,
  toSPSControllerSystemTypeInput,
  updateSPSControllerSystemTypeEntry,
  type SPSControllerSystemTypeEntry
} from './SPSControllerFormDraft.js';

const translate = (key: string) => key;

describe('SPSControllerFormDraft', () => {
  it('builds generated device names from building, cabinet, and GA device', () => {
    const cabinet = { control_cabinet_nr: 'ak01' } as ControlCabinet;
    const building = { iws_code: 'iws' } as Building;

    expect(buildSPSControllerDeviceName(cabinet, building, 'A01')).toBe('IWS_AK01_A01');
    expect(buildSPSControllerDeviceName(cabinet, building, '')).toBe('');
    expect(buildSPSControllerDeviceName(cabinet, null, 'A01')).toBeNull();
  });

  it('formats system type labels with padded number ranges', () => {
    expect(buildSPSControllerSystemTypeLabel('HVAC', 1, 42)).toBe('HVAC (0001-0042)');
  });

  it('updates system type draft fields', () => {
    const entry: SPSControllerSystemTypeEntry = {
      id: 'row-1',
      system_type_id: 'type-1',
      number: 1,
      document_name: 'Old'
    };

    expect(updateSPSControllerSystemTypeEntry(entry, 'number', '7').number).toBe(7);
    expect(updateSPSControllerSystemTypeEntry(entry, 'number', '').number).toBeUndefined();
    expect(
      updateSPSControllerSystemTypeEntry(entry, 'document_name', '').document_name
    ).toBeUndefined();
    expect(updateSPSControllerSystemTypeEntry(entry, 'document_name', 'Doc').document_name).toBe(
      'Doc'
    );
  });

  it('finds next available system type number inside range', () => {
    const entries: SPSControllerSystemTypeEntry[] = [
      { system_type_id: 'type-1', number: 1 },
      { system_type_id: 'type-1', number: 2 },
      { system_type_id: 'type-2', number: 1 }
    ];

    expect(getNextAvailableSystemTypeNumber(entries, 'type-1', 1, 3)).toBe(3);
    expect(getNextAvailableSystemTypeNumber(entries, 'type-1', 1, 2)).toBeNull();
  });

  it('reports add-state from selected system type and used numbers', () => {
    const details = {
      'type-1': {
        id: 'type-1',
        name: 'HVAC',
        number_min: 1,
        number_max: 2
      } as SystemType
    };

    expect(getSPSControllerSystemTypeAddState('', details, [], false, translate)).toEqual({
      disabled: true,
      tooltip: 'facility.forms.sps_controller.system_type_select_first'
    });

    expect(
      getSPSControllerSystemTypeAddState(
        'type-1',
        details,
        [
          { system_type_id: 'type-1', number: 1 },
          { system_type_id: 'type-1', number: 2 }
        ],
        false,
        translate
      )
    ).toEqual({
      disabled: true,
      tooltip: 'facility.forms.sps_controller.system_type_all_used'
    });

    expect(getSPSControllerSystemTypeAddState('type-1', details, [], false, translate)).toEqual({
      disabled: false,
      tooltip: 'facility.forms.sps_controller.system_type_add_next'
    });
  });

  it('maps entries to submit payload without UI-only fields', () => {
    expect(
      toSPSControllerSystemTypeInput([
        {
          id: 'row-1',
          system_type_id: 'type-1',
          number: 1,
          document_name: 'Doc'
        }
      ])
    ).toEqual([
      {
        id: 'row-1',
        system_type_id: 'type-1',
        number: 1,
        document_name: 'Doc'
      }
    ]);
  });
});
