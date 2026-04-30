import type { Specification, SpecificationInput } from '$lib/domain/facility/index.js';

export function buildSpecificationPatch(
  spec: SpecificationInput | undefined
): SpecificationInput | undefined {
  if (!spec) return undefined;

  const patch: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(spec)) {
    if (value !== undefined) {
      patch[key] = value;
    }
  }

  return Object.keys(patch).length > 0 ? (patch as SpecificationInput) : undefined;
}

export function toDisplayOptionalValue<T>(value: T | null | undefined): T | undefined {
  return value === null ? undefined : value;
}

export function normalizeSpecificationForDisplay(
  spec: SpecificationInput | undefined
): Partial<Specification> | undefined {
  if (!spec) return undefined;

  const normalized: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(spec)) {
    normalized[key] = toDisplayOptionalValue(value);
  }
  return normalized as Partial<Specification>;
}
