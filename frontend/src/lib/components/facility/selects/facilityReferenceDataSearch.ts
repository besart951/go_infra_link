import type { Apparat, SystemPart } from '$lib/domain/facility/index.js';

function matchesSearch(search: string, values: Array<string | undefined>): boolean {
  const normalizedSearch = search.trim().toLowerCase();
  if (!normalizedSearch) return true;

  return values.some((value) => value?.toLowerCase().includes(normalizedSearch));
}

export function matchesApparatSearch(apparat: Apparat, search: string): boolean {
  return matchesSearch(search, [apparat.short_name, apparat.name, apparat.description]);
}

export function matchesSystemPartSearch(systemPart: SystemPart, search: string): boolean {
  return matchesSearch(search, [systemPart.short_name, systemPart.name]);
}
