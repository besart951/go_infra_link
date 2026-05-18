import { api } from '$lib/api/client.js';
import type {
  BacnetReferenceResource,
  BacnetReferenceUsageListResponse
} from '$lib/domain/facility/index.js';

export const bacnetReferenceUsageRepository = {
  async getCounts(
    resource: BacnetReferenceResource,
    ids: string[],
    signal?: AbortSignal
  ): Promise<Record<string, number>> {
    const uniqueIds = Array.from(new Set(ids.filter(Boolean)));
    if (uniqueIds.length === 0) return {};

    const params = new URLSearchParams({ resource });
    for (const id of uniqueIds) {
      params.append('ids', id);
    }

    const response = await api<BacnetReferenceUsageListResponse>(
      `/facility/bacnet-reference-usages?${params.toString()}`,
      { signal, skipHttpErrorNavigation: true }
    );

    return Object.fromEntries(
      response.items.map((item) => [item.id, item.bacnet_object_count] as const)
    );
  }
};
