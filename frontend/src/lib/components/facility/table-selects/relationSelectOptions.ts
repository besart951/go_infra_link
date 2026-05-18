export interface RelationSelectOption {
  id: string;
  short_name?: string | null;
  name?: string | null;
}

export interface ApparatSystemPartRelation extends RelationSelectOption {
  system_parts?: RelationSelectOption[];
}

export type RelationFilterSource = 'apparat_id' | 'system_part_id' | null;

export function formatRelationSelectLabel(item: RelationSelectOption): string {
  const shortName = item.short_name?.trim() ?? '';
  const name = item.name?.trim() ?? '';

  if (!shortName) return name;
  if (!name || name === shortName) return shortName;

  return `${shortName} - ${name}`;
}

export function mergeSelectedRelationOption<T extends RelationSelectOption>(
  items: readonly T[],
  selected: T | null | undefined
): T[] {
  if (!selected?.id) return [...items];
  if (items.some((item) => item.id === selected.id)) return [...items];

  return [selected, ...items];
}

export function filterApparatsForSystemPart<T extends ApparatSystemPartRelation>(
  apparats: readonly T[],
  systemPartId: string | null | undefined
): T[] {
  if (!systemPartId) return [...apparats];
  if (!hasLoadedRelationData(apparats)) return [...apparats];

  return apparats.filter((apparat) =>
    (apparat.system_parts ?? []).some((systemPart) => systemPart.id === systemPartId)
  );
}

export function filterApparatsForRelationSource<T extends ApparatSystemPartRelation>(
  apparats: readonly T[],
  systemPartId: string | null | undefined,
  filterSource: RelationFilterSource
): T[] {
  return filterSource === 'system_part_id'
    ? filterApparatsForSystemPart(apparats, systemPartId)
    : [...apparats];
}

export function filterSystemPartsForApparat<T extends RelationSelectOption>(
  systemParts: readonly T[],
  apparats: readonly ApparatSystemPartRelation[],
  apparatId: string | null | undefined
): T[] {
  if (!apparatId) return [...systemParts];

  const allowedIds = getSystemPartIdsForApparat(apparats, apparatId);
  if (!allowedIds) return [...systemParts];

  return systemParts.filter((systemPart) => allowedIds.has(systemPart.id));
}

export function filterSystemPartsForRelationSource<T extends RelationSelectOption>(
  systemParts: readonly T[],
  apparats: readonly ApparatSystemPartRelation[],
  apparatId: string | null | undefined,
  filterSource: RelationFilterSource
): T[] {
  return filterSource === 'apparat_id'
    ? filterSystemPartsForApparat(systemParts, apparats, apparatId)
    : [...systemParts];
}

export function isSystemPartAllowedForApparat(
  apparats: readonly ApparatSystemPartRelation[],
  apparatId: string | null | undefined,
  systemPartId: string | null | undefined
): boolean {
  if (!apparatId || !systemPartId) return true;

  const allowedIds = getSystemPartIdsForApparat(apparats, apparatId);
  if (!allowedIds) return true;

  return allowedIds.has(systemPartId);
}

function getSystemPartIdsForApparat(
  apparats: readonly ApparatSystemPartRelation[],
  apparatId: string
): Set<string> | undefined {
  const apparat = apparats.find((item) => item.id === apparatId);
  if (!apparat || !Array.isArray(apparat.system_parts)) return undefined;

  return new Set(apparat.system_parts.map((systemPart) => systemPart.id));
}

function hasLoadedRelationData(apparats: readonly ApparatSystemPartRelation[]): boolean {
  return apparats.some((apparat) => Array.isArray(apparat.system_parts));
}
