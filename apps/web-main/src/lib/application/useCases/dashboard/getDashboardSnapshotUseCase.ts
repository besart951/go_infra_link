import type { DashboardSnapshot } from '$lib/domain/dashboard/index.js';
import type { DashboardRepository } from '$lib/domain/ports/dashboard/dashboardRepository.js';

export class GetDashboardSnapshotUseCase {
  constructor(private readonly repository: DashboardRepository) {}

  execute(signal?: AbortSignal): Promise<DashboardSnapshot> {
    return this.repository.getSnapshot(signal);
  }
}
