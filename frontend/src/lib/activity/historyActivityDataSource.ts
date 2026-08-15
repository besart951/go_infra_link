import type { ActivityDataSource } from './contract.js';
import { historyRepository } from '$lib/infrastructure/api/historyRepository.js';

export const globalActivityDataSource: ActivityDataSource = {
  cacheNamespace: 'history:global',
  list: (params, signal) => historyRepository.listTimeline(params, signal)
};

export function projectActivityDataSource(projectId: string): ActivityDataSource {
  return {
    cacheNamespace: `history:project:${projectId}`,
    list: (params, signal) => historyRepository.listProjectTimeline(projectId, params, signal)
  };
}
