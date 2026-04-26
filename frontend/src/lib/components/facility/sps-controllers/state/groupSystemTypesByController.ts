import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';

export function groupSystemTypesByController(
  items: SPSControllerSystemType[]
): Record<string, SPSControllerSystemType[]> {
  const grouped: Record<string, SPSControllerSystemType[]> = {};

  for (const item of items) {
    if (!item.sps_controller_id) {
      continue;
    }

    grouped[item.sps_controller_id] = [...(grouped[item.sps_controller_id] ?? []), item];
  }

  return grouped;
}
