import { describe, expect, it } from 'vitest';
import { VERSIONED_FIELD_DEVICE_SHEETS } from '$lib/infrastructure/api/versionedFieldDeviceImport.js';
import { VersionedFieldDeviceImportState } from './VersionedFieldDeviceImportState.svelte.js';

describe('VersionedFieldDeviceImportState', () => {
  it('keeps the server validation result without callbacks', async () => {
    const state = new VersionedFieldDeviceImportState(async () => ({
      import_id: 'import-1',
      total: 2,
      imported: 1,
      failed: 1,
      issues: [{ code: 'aggregate_import_failed', entity: 'field_device', message: 'conflict' }]
    }));
    state.select(new File(['data'], 'facility.xlsx'), VERSIONED_FIELD_DEVICE_SHEETS);

    const result = await state.run();

    expect(result?.imported).toBe(1);
    expect(state.result?.failed).toBe(1);
    expect(state.isImporting).toBe(false);
  });

  it('selects a workbook shard archive without client-side parsing', () => {
    const state = new VersionedFieldDeviceImportState();

    state.selectArchive(new File(['archive'], 'facility.zip'));

    expect(state.isVersioned).toBe(true);
    expect(state.isArchive).toBe(true);
  });
});
