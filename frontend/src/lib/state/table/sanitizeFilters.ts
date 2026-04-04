import type { TableFilterRecord } from './contracts.js';

export function sanitizeFilters<TFilters extends TableFilterRecord>(filters: TFilters): TFilters {
  return Object.fromEntries(
    Object.entries(filters).filter(([, value]) => value !== undefined && value !== '')
  ) as TFilters;
}
