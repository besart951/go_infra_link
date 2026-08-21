import { describe, expect, it } from 'vitest';
import { buildFieldDeviceExportRequest, hasExportScope } from './exportRequest.js';

describe('buildFieldDeviceExportRequest', () => {
  it('maps the current list scope and de-duplicates filter IDs', () => {
    const request = buildFieldDeviceExportRequest({
      projectId: 'project-a',
      search: ' pump ',
      filters: {
        projectId: 'project-a',
        buildingIds: 'building-a, building-b',
        spsControllerSystemTypeId: 'system-type-a'
      }
    });

    expect(request).toEqual({
      project_ids: ['project-a'],
      buildings_id: ['building-a', 'building-b'],
      sps_controller_system_type_ids: ['system-type-a'],
      search: 'pump'
    });
    expect(hasExportScope(request)).toBe(true);
  });

  it('requires explicit export_all for an unfiltered global export', () => {
    expect(hasExportScope(buildFieldDeviceExportRequest({}))).toBe(false);
    expect(hasExportScope({ export_all: true })).toBe(true);
  });
});
