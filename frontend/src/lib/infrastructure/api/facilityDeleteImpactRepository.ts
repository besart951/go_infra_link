import { apiClient } from '$lib/api/generated/client.js';
import type {
  FacilityDeleteImpact,
  FacilityDeleteImpactResource
} from '$lib/domain/facility/index.js';

export const facilityDeleteImpactRepository = {
  async getImpacts(
    resource: FacilityDeleteImpactResource,
    ids: string[],
    signal?: AbortSignal
  ): Promise<FacilityDeleteImpact[]> {
    const uniqueIDs = Array.from(new Set(ids.filter(Boolean)));
    if (uniqueIDs.length === 0) return [];

    const { data } = await apiClient.GET('/api/v1/facility/delete-impacts', {
      params: { query: { resource, ids: uniqueIDs } },
      signal
    });
    if (!data) throw new Error('Die Löschvorschau-Antwort ist leer.');
    return (data.items ?? []).flatMap((impact): FacilityDeleteImpact[] => {
      if (
        typeof impact.id !== 'string' ||
        (impact.resource !== 'apparat' && impact.resource !== 'system_part')
      ) {
        return [];
      }
      return [
        {
          id: impact.id,
          resource: impact.resource,
          blockers: (impact.blockers ?? []).flatMap((blocker) =>
            typeof blocker.resource === 'string' && typeof blocker.count === 'number'
              ? [{ resource: blocker.resource, count: blocker.count }]
              : []
          )
        }
      ];
    });
  }
};
