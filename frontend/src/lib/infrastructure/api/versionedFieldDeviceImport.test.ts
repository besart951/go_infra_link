import { describe, expect, it, vi } from 'vitest';
import {
  isVersionedFieldDeviceArchive,
  isVersionedFieldDeviceWorkbook,
  uploadVersionedFieldDeviceWorkbook,
  VERSIONED_FIELD_DEVICE_SHEETS
} from './versionedFieldDeviceImport.js';

describe('versioned field-device import', () => {
  it('recognizes only complete versioned workbooks', () => {
    expect(isVersionedFieldDeviceWorkbook(VERSIONED_FIELD_DEVICE_SHEETS)).toBe(true);
    expect(isVersionedFieldDeviceWorkbook(['Export-Manifest', 'Data-FieldDevices'])).toBe(false);
  });

  it('recognizes workbook shard archives by extension', () => {
    expect(isVersionedFieldDeviceArchive(new File([], 'facility.ZIP'))).toBe(true);
    expect(isVersionedFieldDeviceArchive(new File([], 'facility.xlsx'))).toBe(false);
  });

  it('uploads a multipart file and preserves a 422 validation report', async () => {
    const transport = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
      expect(init?.body).toBeInstanceOf(FormData);
      return new Response(
        JSON.stringify({
          import_id: 'import-1',
          total: 1,
          imported: 0,
          failed: 0,
          issues: [{ code: 'missing_owner' }]
        }),
        {
          status: 422,
          headers: { 'Content-Type': 'application/json' }
        }
      );
    });

    const result = await uploadVersionedFieldDeviceWorkbook(
      new File(['workbook'], 'facility.xlsx'),
      transport
    );

    expect(result.issues?.[0]?.code).toBe('missing_owner');
    expect(transport).toHaveBeenCalledOnce();
  });
});
