import { getErrorMessage } from '$lib/api/client.js';
import type { HistoryTimelineParams } from '$lib/domain/history.js';
import type { User } from '$lib/domain/user/index.js';
import { historyRepository } from '$lib/infrastructure/api/historyRepository.js';
import { getUser, listUsers } from '$lib/infrastructure/api/user.adapter.js';

export function timelineErrorMessage(error: unknown): string {
  return getErrorMessage(error);
}

export function loadHistoryTimeline(params: HistoryTimelineParams) {
  return historyRepository.listTimeline(params);
}

export function restoreTimelineEvent(eventId: string) {
  return historyRepository.restoreEvent(eventId, 'before');
}

export async function fetchTimelineUsers(search: string): Promise<User[]> {
  try {
    const response = await listUsers({
      page: 1,
      limit: 20,
      search: search.trim() || undefined
    });
    return response.items;
  } catch {
    return [];
  }
}

export async function fetchTimelineUser(id: string): Promise<User | null> {
  try {
    return await getUser(id);
  } catch {
    return null;
  }
}
