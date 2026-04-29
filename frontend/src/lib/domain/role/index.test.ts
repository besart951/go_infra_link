/// <reference types="vitest" />

import { createPermissionName, parsePermissionName } from './index.js';

describe('role permission helpers', () => {
  it('creates nested project permission names', () => {
    expect(createPermissionName('project', 'update', 'fielddevice.bacnetobjects')).toBe(
      'project.fielddevice.bacnetobjects.update'
    );
  });

  it('parses nested project bacnet permissions', () => {
    expect(parsePermissionName('project.fielddevice.bacnetobjects.update')).toEqual({
      resource: 'project',
      subResource: 'fielddevice.bacnetobjects',
      action: 'update',
      category: 'project'
    });
  });

  it('parses underscored project specification permissions', () => {
    expect(parsePermissionName('project.fielddevice_specification.edit')).toEqual({
      resource: 'project',
      subResource: 'fielddevice_specification',
      action: 'edit',
      category: 'project'
    });
  });
});