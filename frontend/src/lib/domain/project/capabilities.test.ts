import { describe, expect, it } from 'vitest';
import { toProjectCapabilities } from './capabilities.js';

describe('project capabilities', () => {
  it('keeps only canonical, project-scoped permissions', () => {
    expect(
      toProjectCapabilities([
        'project.fielddevice.update',
        'project.create',
        'fielddevice.update',
        'project.unknown'
      ]).permissions
    ).toEqual(['project.fielddevice.update']);
  });
});
