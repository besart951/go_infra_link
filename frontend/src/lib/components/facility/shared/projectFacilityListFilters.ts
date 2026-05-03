import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import { sanitizeFilters } from '$lib/state/table/sanitizeFilters.js';

const STORAGE_PREFIX = 'go-infra-link:project-facility-list-filters';
const MULTI_FILTER_SEPARATOR = '|';

export interface MultiFilterOption {
  id: string;
  label: string;
  count?: number;
}

export function normalizeMultiFilterIds(ids: string[]): string[] {
  return [...new Set(ids.map((id) => id.trim()).filter(Boolean))];
}

export function encodeMultiFilter(ids: string[]): string | undefined {
  const normalizedIds = normalizeMultiFilterIds(ids);
  if (normalizedIds.length === 0) return undefined;

  return normalizedIds.map((id) => encodeURIComponent(id)).join(MULTI_FILTER_SEPARATOR);
}

export function decodeMultiFilter(value: string | undefined): string[] {
  if (!value) return [];

  return normalizeMultiFilterIds(
    value
      .split(MULTI_FILTER_SEPARATOR)
      .map((id) => decodeURIComponent(id))
      .filter(Boolean)
  );
}

export function sortMultiFilterOptions(options: MultiFilterOption[]): MultiFilterOption[] {
  return [...options].sort((a, b) => a.label.localeCompare(b.label, undefined, { numeric: true }));
}

export function sanitizeMultiFilterValue(
  value: string | undefined,
  availableIds: Set<string>
): string | undefined {
  const ids = decodeMultiFilter(value).filter((id) => availableIds.has(id));
  return encodeMultiFilter(ids);
}

export class ProjectFacilityListFilterStore<TFilters extends TableFilterRecord> {
  constructor(private readonly scope: string) {}

  load(projectId: string | undefined): TFilters {
    if (!projectId || typeof localStorage === 'undefined') return {} as TFilters;

    try {
      const raw = localStorage.getItem(this.storageKey(projectId));
      if (!raw) return {} as TFilters;

      const parsed = JSON.parse(raw) as Partial<TFilters>;
      return sanitizeFilters(parsed as TFilters);
    } catch {
      return {} as TFilters;
    }
  }

  save(projectId: string | undefined, filters: TFilters): void {
    if (!projectId || typeof localStorage === 'undefined') return;

    const cleanFilters = sanitizeFilters(filters);
    const key = this.storageKey(projectId);

    try {
      if (Object.keys(cleanFilters).length === 0) {
        localStorage.removeItem(key);
        return;
      }

      localStorage.setItem(key, JSON.stringify(cleanFilters));
    } catch {
      // Storage failures must not break list filtering.
    }
  }

  private storageKey(projectId: string): string {
    return `${STORAGE_PREFIX}:${this.scope}:${projectId}`;
  }
}
