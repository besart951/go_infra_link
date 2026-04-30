import { describe, expect, it } from 'vitest';
import {
  buildMultiCreatePayload,
  collectCreatedDevices,
  normalizeMultiCreateResponse,
  reconcileMultiCreateRows
} from './multiCreateSubmission.js';

const rows = [
  { id: 'row-1', bmk: 'BMK-1', description: '', textFix: null, apparatNr: 11 },
  { id: 'row-2', bmk: 'BMK-2', description: 'Desc', textFix: 'TF', apparatNr: 12 }
];

describe('multi-create submission helpers', () => {
  it('builds create payload from rows and selection', () => {
    expect(
      buildMultiCreatePayload(rows, {
        spsControllerSystemTypeId: 'sps-1',
        objectDataId: 'obj-1',
        apparatId: 'app-1',
        systemPartId: 'sp-1'
      })
    ).toEqual([
      {
        bmk: 'BMK-1',
        description: undefined,
        text_fix: undefined,
        apparat_nr: 11,
        sps_controller_system_type_id: 'sps-1',
        system_part_id: 'sp-1',
        apparat_id: 'app-1',
        object_data_id: 'obj-1'
      },
      {
        bmk: 'BMK-2',
        description: 'Desc',
        text_fix: 'TF',
        apparat_nr: 12,
        sps_controller_system_type_id: 'sps-1',
        system_part_id: 'sp-1',
        apparat_id: 'app-1',
        object_data_id: 'obj-1'
      }
    ]);
  });

  it('normalizes wrapped preview responses', () => {
    const response = {
      results: [],
      total_requests: 0,
      success_count: 0,
      failure_count: 0
    };
    expect(normalizeMultiCreateResponse({ preview: response }, 'fallback')).toBe(response);
    expect(() => normalizeMultiCreateResponse({}, 'fallback')).toThrow('fallback');
  });

  it('removes successful rows and remaps failed backend errors', () => {
    const result = reconcileMultiCreateRows(
      rows,
      {
        results: [
          { index: 0, success: true, error: '', error_field: '' },
          { index: 1, success: false, error: 'bad', error_field: 'apparat_nr' }
        ],
        total_requests: 2,
        success_count: 1,
        failure_count: 1
      },
      (message, field) => `${field}:${message}`
    );

    expect(result.remainingRows).toEqual([rows[1]]);
    expect(result.rowErrors.get(0)).toEqual({
      message: 'apparat_nr:bad',
      field: 'apparat_nr'
    });
  });

  it('collects created devices', () => {
    expect(
      collectCreatedDevices({
        results: [
          {
            index: 0,
            success: true,
            error: '',
            error_field: '',
            field_device: { id: 'fd-1' } as never
          }
        ],
        total_requests: 1,
        success_count: 1,
        failure_count: 0
      })
    ).toEqual([{ id: 'fd-1' }]);
  });
});
