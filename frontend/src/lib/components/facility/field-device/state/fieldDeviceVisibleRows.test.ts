import { describe, expect, it } from 'vitest';
import { buildFieldDevice } from '$lib/test/fieldDevice.fixtures.js';
import {
  applySPSControllerNameDelta,
  planSPSControllerDeviceRefresh,
  planVisibleDeviceDelta,
  planVisibleDeviceRefresh
} from './fieldDeviceVisibleRows.js';

const visibleDevice = buildFieldDevice({
  id: 'fd-1',
  sps_controller_system_type: {
    id: 'sst-1',
    system_type_id: 'st-1',
    sps_controller_id: 'sps-1',
    sps_controller_name: 'SPS old',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  }
});

const context = {
  items: [visibleDevice],
  hasActiveFilters: false
};

describe('field-device visible row helpers', () => {
  it('plans targeted refresh only for visible ids without query state', () => {
    expect(planVisibleDeviceRefresh(context, ['fd-1', 'fd-1'])).toEqual({
      action: 'refresh',
      ids: ['fd-1']
    });
    expect(planVisibleDeviceRefresh(context, ['fd-2'])).toEqual({ action: 'reload' });
    expect(planVisibleDeviceRefresh({ ...context, searchText: 'pump' }, ['fd-1'])).toEqual({
      action: 'reload'
    });
  });

  it('plans visible deltas without full reload when all changed devices are visible', () => {
    expect(planVisibleDeviceDelta(context, [{ ...visibleDevice, bmk: 'FD-UPDATED' }])).toEqual({
      action: 'replace',
      devices: [{ ...visibleDevice, bmk: 'FD-UPDATED' }]
    });
    expect(planVisibleDeviceDelta(context, [buildFieldDevice({ id: 'fd-2' })])).toEqual({
      action: 'reload'
    });
  });

  it('maps SPS controller refreshes to visible device ids', () => {
    expect(planSPSControllerDeviceRefresh(context, ['sps-1'])).toEqual({
      action: 'refresh',
      ids: ['fd-1']
    });
    expect(planSPSControllerDeviceRefresh(context, ['sps-2'])).toEqual({ action: 'none' });
  });

  it('applies SPS controller name deltas to visible field devices', () => {
    expect(
      applySPSControllerNameDelta(context.items, [
        {
          id: 'sps-1',
          device_name: 'SPS new',
          control_cabinet_id: 'cc-1',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ])[0].sps_controller_system_type?.sps_controller_name
    ).toBe('SPS new');
  });
});
