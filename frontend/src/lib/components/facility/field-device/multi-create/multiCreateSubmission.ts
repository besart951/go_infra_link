import type {
  CreateFieldDeviceRequest,
  FieldDevice,
  MultiCreateFieldDeviceResponse
} from '$lib/domain/facility/index.js';
import type {
  FieldDeviceRowData,
  FieldDeviceRowError,
  MultiCreateSelection
} from '$lib/domain/facility/fieldDeviceMultiCreate.js';

export function buildMultiCreatePayload(
  rows: FieldDeviceRowData[],
  selection: MultiCreateSelection
): CreateFieldDeviceRequest[] {
  return rows.map((row) => ({
    bmk: row.bmk || undefined,
    description: row.description || undefined,
    text_fix: row.textFix || undefined,
    apparat_nr: row.apparatNr!,
    sps_controller_system_type_id: selection.spsControllerSystemTypeId,
    system_part_id: selection.systemPartId,
    apparat_id: selection.apparatId,
    object_data_id: selection.objectDataId || undefined
  }));
}

export function normalizeMultiCreateResponse(
  response: MultiCreateFieldDeviceResponse | { preview?: MultiCreateFieldDeviceResponse },
  fallbackMessage: string
): MultiCreateFieldDeviceResponse {
  if (isMultiCreateFieldDeviceResponse(response)) {
    return response;
  }

  if (
    'preview' in response &&
    response.preview &&
    isMultiCreateFieldDeviceResponse(response.preview)
  ) {
    return response.preview;
  }

  throw new Error(fallbackMessage);
}

export function reconcileMultiCreateRows(
  rows: FieldDeviceRowData[],
  response: MultiCreateFieldDeviceResponse,
  localizeError: (message: string, field?: string) => string
): {
  remainingRows: FieldDeviceRowData[];
  rowErrors: Map<number, FieldDeviceRowError>;
} {
  const successfulIndices = new Set<number>();
  const backendErrors = new Map<number, FieldDeviceRowError>();

  for (const result of response.results) {
    if (result.index < 0 || result.index >= rows.length) {
      continue;
    }

    if (result.success) {
      successfulIndices.add(result.index);
      continue;
    }

    backendErrors.set(result.index, {
      message: localizeError(result.error, result.error_field),
      field: (result.error_field as FieldDeviceRowError['field']) || ''
    });
  }

  const indexMap = new Map<number, number>();
  const remainingRows: FieldDeviceRowData[] = [];

  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index];
    if (!successfulIndices.has(index)) {
      indexMap.set(index, remainingRows.length);
      remainingRows.push(row);
    }
  }

  const rowErrors = new Map<number, FieldDeviceRowError>();
  for (const [originalIndex, error] of backendErrors.entries()) {
    const nextIndex = indexMap.get(originalIndex);
    if (nextIndex !== undefined) {
      rowErrors.set(nextIndex, error);
    }
  }

  return { remainingRows, rowErrors };
}

export function collectCreatedDevices(response: MultiCreateFieldDeviceResponse): FieldDevice[] {
  return response.results
    .filter((result) => result.success && result.field_device)
    .map((result) => result.field_device!);
}

function isMultiCreateFieldDeviceResponse(value: unknown): value is MultiCreateFieldDeviceResponse {
  if (!value || typeof value !== 'object') {
    return false;
  }

  const candidate = value as Record<string, unknown>;

  return (
    Array.isArray(candidate.results) &&
    typeof candidate.total_requests === 'number' &&
    typeof candidate.success_count === 'number' &&
    typeof candidate.failure_count === 'number'
  );
}
