import { api, type ApiOptions } from '$lib/api/client.js';
import type { DashboardSnapshot } from '$lib/domain/dashboard/index.js';
import type { DashboardRepository } from '$lib/domain/ports/dashboard/dashboardRepository.js';

export async function getDashboardSnapshot(options?: ApiOptions): Promise<DashboardSnapshot> {
  return api<DashboardSnapshot>('/dashboard', options);
}

export class DashboardApiRepository implements DashboardRepository {
  async getSnapshot(signal?: AbortSignal): Promise<DashboardSnapshot> {
    return getDashboardSnapshot({ signal });
  }
}
