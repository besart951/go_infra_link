/// <reference types="vitest" />

import { beforeEach, describe, expect, it, vi } from 'vitest';

const realtime = vi.hoisted(() => {
  let progressListener: ((event: Record<string, unknown>) => void) | undefined;
  return {
    emit(event: Record<string, unknown>) {
      progressListener?.(event);
    },
    cache: {
      subscribeJobProgress(listener: (event: Record<string, unknown>) => void) {
        progressListener = listener;
        return () => (progressListener = undefined);
      },
      subscribeRealtimeOpen() {
        return () => undefined;
      }
    }
  };
});

const getJob = vi.hoisted(() => vi.fn());
const listJobs = vi.hoisted(() => vi.fn());

vi.mock('$lib/services/facilityReferenceDataCache.js', () => ({
  facilityReferenceDataCache: realtime.cache
}));
vi.mock('$lib/infrastructure/api/facilityJobRepository.js', () => ({
  facilityJobRepository: { get: getJob, list: listJobs, retry: vi.fn() }
}));

import { FacilityJobState } from './facilityJobState.svelte.js';

const jobId = 'c7c1eaa6-21bb-4a2a-b0b7-2d809f313018';

describe('FacilityJobState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    listJobs.mockResolvedValue({ items: [] });
    sessionStorage.clear();
    vi.stubGlobal('crypto', { randomUUID: () => jobId });
  });

  it('recovers active jobs from the backend without a copy session key', async () => {
    listJobs.mockResolvedValue({
      items: [
        {
          jobId,
          kind: 'object_data',
          type: 'copy',
          status: 'running',
          progress: 37,
          stage: 'copying_root'
        }
      ]
    });
    const state = new FacilityJobState();
    state.initialize('user-a');
    await vi.waitFor(() => expect(state.isPending).toBe(true));
    expect(state.progress).toBe(37);
    expect(sessionStorage.length).toBe(0);
    state.dispose();
  });

  it('submits once and settles from the canonical job event', async () => {
    const state = new FacilityJobState();
    const start = vi.fn(async () => ({
      jobId,
      kind: 'sps_controller' as const,
      type: 'copy' as const,
      status: 'running' as const,
      progress: 1,
      stage: 'preparing'
    }));
    await expect(state.submit(start)).resolves.toEqual({ started: true });
    await expect(state.submit(start)).resolves.toEqual({ started: false });
    realtime.emit({
      type: 'facility.job.progress',
      job_id: jobId,
      kind: 'sps_controller',
      job_type: 'copy',
      status: 'completed',
      progress: 100,
      stage: 'completed',
      updated_at: new Date().toISOString()
    });
    await vi.waitFor(() => expect(state.isPending).toBe(false));
    expect(start).toHaveBeenCalledOnce();
  });
});
