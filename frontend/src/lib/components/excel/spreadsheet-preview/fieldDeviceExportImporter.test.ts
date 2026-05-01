import { describe, expect, it } from 'vitest';
import { WorksheetPreview } from '$lib/domain/excel/index.js';
import type {
  AlarmType,
  Apparat,
  Building,
  ControlCabinet,
  FieldDevice,
  NotificationClass,
  StateText,
  SystemPart,
  SystemType
} from '$lib/domain/facility/index.js';
import {
  buildImportCellMarkers,
  collectFieldDeviceImportLookupCriteria,
  transformWorksheetToFieldDeviceImport,
  type FieldDeviceImportLookups
} from './fieldDeviceExportImporter.js';

type TestCell = string | number | boolean;

function row(values: Record<number, TestCell>): TestCell[] {
  const out: TestCell[] = Array.from({ length: 45 }, () => '');
  for (const [index, value] of Object.entries(values)) {
    out[Number(index)] = value;
  }
  return out;
}

function worksheet(rows: TestCell[][]): WorksheetPreview {
  return new WorksheetPreview('Projekt Controller ABC', rows);
}

function baseRows(): TestCell[][] {
  return [
    row({ 0: 'Projekt Controller', 1: 'ABC' }),
    row({ 0: 'GA-Gerät:', 1: 'ABC' }),
    row({ 0: 'Schaltschrank-Nr.', 1: '1_0100_00' }),
    row({ 0: 'Device Name:', 1: 'IWS_1_0100_ABC' }),
    row({ 0: 'Device Instance:', 1: 'WS001' }),
    row({ 0: 'Device Description:', 1: 'Controller description' }),
    row({ 0: 'Device Location:', 1: 'Technikraum' }),
    row({ 0: 'IP-Adresse:', 1: '192.168.1.10' }),
    row({ 0: 'Subnetz:', 1: '255.255.255.0' }),
    row({ 0: 'Gateway:', 1: '192.168.1.1' }),
    row({ 0: 'VLAN:', 1: '10' }),
    row({}),
    row({
      0: 'BACnet Object Name',
      1: 'Description',
      2: 'State Text (Tabelle der Zustandstexte) / (B2) /',
      3: 'Notification',
      4: 'BMK'
    }),
    row({
      0: 'IWS_1_0100_ABC_HVPMP01',
      1: 'Heating Pump',
      4: 'B001',
      5: true,
      6: 'Heating',
      7: 'Pump',
      8: 'HV',
      9: 'PMP',
      33: 'Field device remark'
    }),
    row({
      0: 'IWS_1_0100_ABC_HVPMP01_AI01',
      1: 'Heating Pump - HVPMP Status',
      2: 'Off, On',
      3: 7,
      4: 'B001',
      5: true,
      6: 'Heating',
      7: 'Pump',
      8: 'HV',
      9: 'PMP',
      10: 'Status',
      11: 1,
      26: 'AI01',
      27: 'Alarm',
      28: 'Off',
      32: 1
    })
  ];
}

function lookups(overrides: Partial<FieldDeviceImportLookups> = {}): FieldDeviceImportLookups {
  const building: Building = {
    id: 'building-1',
    iws_code: 'IWS',
    building_group: 1,
    created_at: '',
    updated_at: ''
  };
  const systemType: SystemType = {
    id: 'system-type-1',
    number_min: 100,
    number_max: 199,
    name: 'Lüftung',
    created_at: '',
    updated_at: ''
  };
  const systemPart: SystemPart = {
    id: 'system-part-1',
    short_name: 'HV',
    name: 'Heating',
    created_at: '',
    updated_at: ''
  };
  const apparat: Apparat = {
    id: 'apparat-1',
    short_name: 'PMP',
    name: 'Pump',
    system_parts: [systemPart],
    created_at: '',
    updated_at: ''
  };
  const stateText: StateText = {
    id: 'state-text-1',
    ref_number: 1,
    state_text1: 'Off',
    state_text2: 'On',
    created_at: '',
    updated_at: ''
  };
  const notificationClass: NotificationClass = {
    id: 'notification-1',
    event_category: '',
    nc: 7,
    object_description: '',
    internal_description: '',
    meaning: '',
    ack_required_not_normal: false,
    ack_required_error: false,
    ack_required_normal: false,
    norm_not_normal: 0,
    norm_error: 0,
    norm_normal: 0,
    created_at: '',
    updated_at: ''
  };
  const alarmType: AlarmType = {
    id: 'alarm-type-1',
    code: 'alarm',
    name: 'Alarm',
    created_at: '',
    updated_at: ''
  };

  return {
    buildings: [building],
    controlCabinets: [],
    spsControllers: [],
    spsControllerSystemTypes: [],
    fieldDevices: [],
    systemTypes: [systemType],
    systemParts: [systemPart],
    apparats: [apparat],
    stateTexts: [stateText],
    notificationClasses: [notificationClass],
    alarmTypes: [alarmType],
    ...overrides
  };
}

describe('field device export importer', () => {
  it('collects lookup criteria from the worksheet before validation lookups are loaded', () => {
    const criteria = collectFieldDeviceImportLookupCriteria(worksheet(baseRows()));

    expect(criteria).toEqual({
      iwsCode: 'IWS',
      buildingGroup: 1,
      controlCabinetNr: '1_0100_00',
      gaDevice: 'ABC',
      deviceName: 'IWS_1_0100_ABC',
      systemTypeNumbers: [100],
      systemPartLabels: ['Heating', 'HV'],
      apparatLabels: ['PMP', 'Pump'],
      notificationNumbers: [7],
      stateTextLabels: ['Off', 'Off, On'],
      alarmTypeLabels: ['Alarm']
    });
  });

  it('transforms an exported worksheet into a controller tree and create payloads', () => {
    const plan = transformWorksheetToFieldDeviceImport(worksheet(baseRows()), lookups());

    expect(plan.canImport).toBe(true);
    expect(plan.errorCount).toBe(0);
    expect(plan.warningCount).toBe(0);
    expect(plan.controller.controlCabinetRequest).toEqual({
      building_id: 'building-1',
      control_cabinet_nr: '1_0100_00'
    });
    expect(plan.controller.spsControllerRequest?.system_types).toEqual([
      { system_type_id: 'system-type-1', number: 100 }
    ]);
    expect(plan.controller.fieldDevices).toHaveLength(1);
    expect(plan.controller.fieldDevices[0].request).toMatchObject({
      apparat_nr: 1,
      system_part_id: 'system-part-1',
      apparat_id: 'apparat-1'
    });
    expect(plan.controller.fieldDevices[0].bacnetObjects[0].request).toMatchObject({
      text_fix: 'Status',
      software_type: 'ai',
      software_number: 1,
      hardware_type: 'ai',
      hardware_quantity: 1,
      notification_class_id: 'notification-1',
      state_text_id: 'state-text-1',
      alarm_type_id: 'alarm-type-1'
    });
  });

  it('marks the exact cell for missing required lookup values', () => {
    const plan = transformWorksheetToFieldDeviceImport(
      worksheet(baseRows()),
      lookups({ apparats: [] })
    );

    expect(plan.canImport).toBe(false);
    expect(plan.errorCount).toBe(1);
    expect(plan.diagnostics[0]).toMatchObject({
      severity: 'error',
      cell: { address: 'J14' }
    });

    const markers = buildImportCellMarkers(plan.diagnostics);
    expect(markers['14:9'].severity).toBe('error');
  });

  it('marks existing cabinets as reusable on the header source cell', () => {
    const cabinet: ControlCabinet = {
      id: 'cabinet-1',
      building_id: 'building-1',
      control_cabinet_nr: '1_0100_00',
      created_at: '',
      updated_at: ''
    };
    const plan = transformWorksheetToFieldDeviceImport(
      worksheet(baseRows()),
      lookups({ controlCabinets: [cabinet] })
    );

    expect(plan.canImport).toBe(true);
    expect(plan.diagnostics).toContainEqual(
      expect.objectContaining({
        severity: 'warning',
        entity: 'control_cabinet',
        cell: expect.objectContaining({ address: 'B3' })
      })
    );
    expect(plan.controller.existingControlCabinet?.id).toBe('cabinet-1');
  });

  it('marks already existing field devices as reusable on the source object name cell', () => {
    const existingFieldDevice: FieldDevice = {
      id: 'field-device-1',
      apparat_nr: '1',
      sps_controller_system_type_id: 'sps-controller-system-type-1',
      system_part_id: 'system-part-1',
      apparat_id: 'apparat-1',
      created_at: '',
      updated_at: '',
      sps_controller_system_type: {
        id: 'sps-controller-system-type-1',
        sps_controller_id: 'sps-controller-1',
        system_type_id: 'system-type-1',
        sps_controller_name: 'IWS_1_0100_ABC',
        system_type_name: 'Lüftung',
        number: 100,
        created_at: '',
        updated_at: ''
      }
    };
    const plan = transformWorksheetToFieldDeviceImport(
      worksheet(baseRows()),
      lookups({ fieldDevices: [existingFieldDevice] })
    );

    expect(plan.canImport).toBe(true);
    expect(plan.controller.fieldDevices[0].existingFieldDeviceId).toBe('field-device-1');
    expect(plan.diagnostics).toContainEqual(
      expect.objectContaining({
        severity: 'warning',
        entity: 'field_device',
        cell: expect.objectContaining({ address: 'A14' }),
        message: expect.stringContaining('existiert bereits')
      })
    );
  });
});
