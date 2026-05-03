import type {
  Building,
  ControlCabinet,
  SPSControllerSystemType,
  SPSControllerSystemTypeInput,
  SystemType
} from '$lib/domain/facility/index.js';

export type SPSControllerSystemTypeEntry = SPSControllerSystemTypeInput & { id?: string };

export interface SystemTypeAddState {
  disabled: boolean;
  tooltip: string;
}

export function formatSPSControllerNumber(value: number): string {
  return String(value).padStart(4, '0');
}

export function buildSPSControllerSystemTypeLabel(name: string, min: number, max: number): string {
  return `${name} (${formatSPSControllerNumber(min)}-${formatSPSControllerNumber(max)})`;
}

export function buildSPSControllerDeviceName(
  cabinet: ControlCabinet | null,
  building: Building | null,
  gaDevice: string
): string | null {
  const ga = gaDevice.trim();
  if (!ga) return '';

  const iwsCode = building?.iws_code?.trim();
  const cabinetNr = cabinet?.control_cabinet_nr?.trim();
  if (!iwsCode || !cabinetNr) return null;

  return `${iwsCode}_${cabinetNr}_${ga}`.toUpperCase();
}

export function collectUniqueSystemTypeIds(items: SPSControllerSystemType[]): string[] {
  return Array.from(new Set(items.map((item) => item.system_type_id).filter(Boolean)));
}

export function collectSystemTypeLabelFallbacks(
  items: SPSControllerSystemType[]
): Record<string, string> {
  const fallbacks: Record<string, string> = {};
  items.forEach((item) => {
    if (item.system_type_name) {
      fallbacks[item.system_type_id] = item.system_type_name;
    }
  });
  return fallbacks;
}

export function toSPSControllerSystemTypeEntries(
  items: SPSControllerSystemType[]
): SPSControllerSystemTypeEntry[] {
  return items.map((item) => ({
    id: item.id,
    system_type_id: item.system_type_id,
    number: item.number ?? undefined,
    document_name: item.document_name ?? undefined
  }));
}

export function toSPSControllerSystemTypeInput(
  entries: SPSControllerSystemTypeEntry[]
): SPSControllerSystemTypeInput[] {
  return entries.map(({ id, system_type_id, number, document_name }) => ({
    id,
    system_type_id,
    number,
    document_name
  }));
}

export function updateSPSControllerSystemTypeEntry(
  entry: SPSControllerSystemTypeEntry,
  field: keyof SPSControllerSystemTypeInput,
  value: string
): SPSControllerSystemTypeEntry {
  if (field === 'number') {
    const parsed = value === '' ? undefined : Number(value);
    return { ...entry, number: Number.isNaN(parsed) ? undefined : parsed };
  }

  return { ...entry, document_name: value || undefined };
}

export function getNextAvailableSystemTypeNumber(
  entries: SPSControllerSystemTypeEntry[],
  systemTypeId: string,
  min: number,
  max: number
): number | null {
  const usedNumbers = new Set(
    entries
      .filter((item) => item.system_type_id === systemTypeId)
      .map((item) => item.number)
      .filter((value): value is number => typeof value === 'number')
  );

  for (let number = min; number <= max; number += 1) {
    if (!usedNumbers.has(number)) return number;
  }

  return null;
}

export function getSPSControllerSystemTypeAddState(
  selectedSystemTypeId: string,
  systemTypeDetails: Record<string, SystemType>,
  entries: SPSControllerSystemTypeEntry[],
  detailsLoading: boolean,
  translate: (key: string) => string
): SystemTypeAddState {
  if (!selectedSystemTypeId) {
    return {
      disabled: true,
      tooltip: translate('facility.forms.sps_controller.system_type_select_first')
    };
  }

  const details = systemTypeDetails[selectedSystemTypeId];
  if (!details) {
    return {
      disabled: true,
      tooltip: detailsLoading
        ? translate('facility.forms.sps_controller.system_type_loading_details')
        : translate('facility.forms.sps_controller.system_type_loading')
    };
  }

  const rangeSize = details.number_max - details.number_min + 1;
  const usedNumbers = new Set(
    entries
      .filter(
        (item) =>
          item.system_type_id === selectedSystemTypeId &&
          typeof item.number === 'number' &&
          item.number >= details.number_min &&
          item.number <= details.number_max
      )
      .map((item) => item.number as number)
  );

  if (usedNumbers.size >= rangeSize) {
    return {
      disabled: true,
      tooltip: translate('facility.forms.sps_controller.system_type_all_used')
    };
  }

  return {
    disabled: false,
    tooltip: translate('facility.forms.sps_controller.system_type_add_next')
  };
}
