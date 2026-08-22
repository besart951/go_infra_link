import { describe, expect, it } from 'vitest';
import { versionedDeletePath, versionedProjectLinkDeletePath } from './versionedMutation.js';

describe('versioned mutation URLs', () => {
  it('always includes the required aggregate version', () => {
    expect(
      versionedDeletePath('/api/v1/facility/buildings', { id: 'building-1', base_version: 7 })
    ).toBe('/api/v1/facility/buildings/building-1?base_version=7');
  });

  it('escapes resource IDs without weakening the query contract', () => {
    expect(
      versionedDeletePath('/api/v1/projects/project-1/field-devices', {
        id: 'link/one',
        base_version: 3
      })
    ).toBe('/api/v1/projects/project-1/field-devices/link%2Fone?base_version=3');
  });

  it('builds project-link deletes from one canonical formatter', () => {
    expect(
      versionedProjectLinkDeletePath('field-devices', {
        project_id: 'project/one',
        link_id: 'link/one',
        base_version: 3
      })
    ).toBe('/projects/project%2Fone/field-devices/link%2Fone?base_version=3');
  });
});
