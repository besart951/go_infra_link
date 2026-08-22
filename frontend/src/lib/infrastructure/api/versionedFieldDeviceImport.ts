import { apiFetch, assertApiSuccess } from '$lib/api/client.js';
import type { components } from '$lib/api/generated/schema.js';

export type FieldDeviceImportIssue =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_application_fielddeviceimport.Issue'];
export type FieldDeviceImportResult =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_application_fielddeviceimport.Result'];

export const VERSIONED_FIELD_DEVICE_SHEETS = [
  'Export-Manifest',
  'Data-FieldDevices',
  'Data-Specifications',
  'Data-BACnetObjects',
  'Data-SoftwareReferences',
  'Data-AlarmValues'
] as const;

export function isVersionedFieldDeviceWorkbook(sheetNames: readonly string[]): boolean {
  const available = new Set(sheetNames);
  return VERSIONED_FIELD_DEVICE_SHEETS.every((name) => available.has(name));
}

export function isVersionedFieldDeviceArchive(file: File): boolean {
  return file.name.toLowerCase().endsWith('.zip');
}

export async function uploadVersionedFieldDeviceWorkbook(
  file: File,
  transport: typeof fetch = apiFetch
): Promise<FieldDeviceImportResult> {
  const body = new FormData();
  body.set('file', file);
  const response = await transport('/api/v1/facility/imports/field-devices', {
    method: 'POST',
    body
  });
  if (response.status !== 422) {
    await assertApiSuccess(response);
  }
  return (await response.json()) as FieldDeviceImportResult;
}
