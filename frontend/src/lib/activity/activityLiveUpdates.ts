import { invalidateActivityCache } from './activityCache.js';

type ProjectActivityListener = () => void;

const projectListeners = new Map<string, Set<ProjectActivityListener>>();

/**
 * Receives an already validated project websocket change. The websocket is
 * intentionally only an invalidation signal: visible history always comes
 * from the audited HTTP timeline endpoint.
 */
export function notifyProjectActivityChanged(projectId: string): void {
  if (!projectId) return;

  invalidateActivityCache(`history:project:${projectId}`);
  invalidateActivityCache('history:global');

  for (const listener of projectListeners.get(projectId) ?? []) {
    listener();
  }
}

export function subscribeToProjectActivity(
  projectId: string,
  listener: ProjectActivityListener
): () => void {
  if (!projectId) return () => undefined;

  const listeners = projectListeners.get(projectId) ?? new Set<ProjectActivityListener>();
  listeners.add(listener);
  projectListeners.set(projectId, listeners);

  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) projectListeners.delete(projectId);
  };
}

export function clearActivityLiveUpdatesForTests(): void {
  projectListeners.clear();
}
