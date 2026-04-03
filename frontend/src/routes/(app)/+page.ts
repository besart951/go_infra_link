import { GetDashboardSnapshotUseCase } from '$lib/application/useCases/dashboard/getDashboardSnapshotUseCase.js';
import { getDashboardSnapshot } from '$lib/infrastructure/api/dashboard.adapter.js';
import { t } from '$lib/i18n/index.js';

export const load = async ({ fetch }: { fetch: typeof globalThis.fetch }) => {
  const useCase = new GetDashboardSnapshotUseCase({
    getSnapshot: () => getDashboardSnapshot({ customFetch: fetch })
  });

  try {
    const dashboard = await useCase.execute();
    return { dashboard, loadError: null };
  } catch (error) {
    const message = error instanceof Error ? error.message : t('dashboard.unavailable');
    return { dashboard: null, loadError: message };
  }
};
