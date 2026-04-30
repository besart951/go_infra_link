import { api } from '$lib/api/client.js';
import type { BacnetObject, UpdateBacnetObjectRequest } from '$lib/domain/facility/index.js';

export function updateBacnetObject(
  id: string,
  data: UpdateBacnetObjectRequest,
  signal?: AbortSignal
): Promise<BacnetObject> {
  return api<BacnetObject>(`/facility/bacnet-objects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
    signal
  });
}
