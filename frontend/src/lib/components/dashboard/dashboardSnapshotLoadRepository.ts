import { getDashboardSnapshot } from '$lib/infrastructure/api/dashboard.adapter.js';

export function createDashboardSnapshotLoadRepository(customFetch: typeof globalThis.fetch) {
  return {
    getSnapshot: () => getDashboardSnapshot({ customFetch })
  };
}
