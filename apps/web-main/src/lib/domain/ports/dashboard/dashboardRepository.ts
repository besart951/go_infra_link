import type { DashboardSnapshot } from '$lib/domain/dashboard/index.js';

export interface DashboardRepository {
  getSnapshot(signal?: AbortSignal): Promise<DashboardSnapshot>;
}
