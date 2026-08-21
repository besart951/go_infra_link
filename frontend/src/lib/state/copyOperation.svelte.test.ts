/// <reference types="vitest" />

import { beforeEach, describe, expect, it, vi } from 'vitest';

const realtime = vi.hoisted(() => {
  let progressListener: ((event: Record<string, unknown>) => void) | undefined;
  let openListener: (() => void) | undefined;

  return {
    emitProgress(event: Record<string, unknown>) {
      progressListener?.(event);
    },
    emitOpen() {
      openListener?.();
    },
    facilityReferenceDataCache: {
      subscribeCopyJobProgress(listener: (event: Record<string, unknown>) => void) {
        progressListener = listener;
        return () => {
          progressListener = undefined;
        };
      },
      subscribeRealtimeOpen(listener: () => void) {
        openListener = listener;
        return () => {
          openListener = undefined;
        };
      }
    }
  };
});

const getCopyJob = vi.hoisted(() => vi.fn());
const listCopyJobs = vi.hoisted(() => vi.fn());

vi.mock('$lib/services/facilityReferenceDataCache.js', () => ({
  facilityReferenceDataCache: realtime.facilityReferenceDataCache
}));
vi.mock('$lib/infrastructure/api/copyJobRepository.js', () => ({
  copyJobRepository: { get: getCopyJob, list: listCopyJobs }
}));

import { CopyOperation } from './copyOperation.svelte.js';

const jobId = 'c7c1eaa6-21bb-4a2a-b0b7-2d809f313018';

describe('CopyOperation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    listCopyJobs.mockResolvedValue([]);
    sessionStorage.clear();
    vi.stubGlobal('crypto', { randomUUID: () => jobId });
  });

  it('recovers an active job from the server after login without session state', async () => {
    listCopyJobs.mockResolvedValue([
      {
        jobId,
        kind: 'object_data',
        status: 'running',
        progress: 37,
        stage: 'copying_root'
      }
    ]);
    const operation = new CopyOperation();

    operation.initialize('user-a');

    await vi.waitFor(() => expect(operation.isPending).toBe(true));
    expect(operation.progress).toBe(37);
    expect(operation.jobs).toHaveLength(1);
    expect(sessionStorage.getItem('facility-copy-job')).toContain(jobId);
    operation.dispose();
  });

  it('tracks websocket progress and keeps duplicate submissions disabled until completion', async () => {
    const operation = new CopyOperation();
    const onCompleted = vi.fn();
    const start = vi.fn(async () => ({
      jobId,
      kind: 'sps_controller' as const,
      status: 'running' as const,
      progress: 1,
      stage: 'preparing'
    }));

    await expect(operation.run({ start, onCompleted, onFailed: vi.fn() })).resolves.toEqual({
      started: true
    });
    expect(operation.isPending).toBe(true);

    await expect(operation.run({ start, onCompleted, onFailed: vi.fn() })).resolves.toEqual({
      started: false
    });
    expect(start).toHaveBeenCalledOnce();

    realtime.emitProgress({
      type: 'facility.copy_job.progress',
      job_id: jobId,
      kind: 'sps_controller',
      status: 'running',
      progress: 63,
      stage: 'copying_field_devices',
      updated_at: new Date().toISOString()
    });
    expect(operation.progress).toBe(63);

    realtime.emitProgress({
      type: 'facility.copy_job.progress',
      job_id: jobId,
      kind: 'sps_controller',
      status: 'completed',
      progress: 100,
      stage: 'completed',
      updated_at: new Date().toISOString()
    });
    await vi.waitFor(() => expect(onCompleted).toHaveBeenCalledOnce());
    expect(operation.isPending).toBe(false);
    expect(sessionStorage.getItem('facility-copy-job')).toBeNull();
  });

  it('reconciles its persisted job after a realtime reconnect', async () => {
    sessionStorage.setItem('facility-copy-job', JSON.stringify({ jobId, ownerId: 'user-a' }));
    getCopyJob.mockResolvedValue({
      jobId,
      kind: 'control_cabinet',
      status: 'running',
      progress: 42,
      stage: 'copying_controllers'
    });
    const operation = new CopyOperation();
    operation.initialize('user-a');

    await vi.waitFor(() => expect(getCopyJob).toHaveBeenCalledWith(jobId));
    expect(operation.isPending).toBe(true);
    expect(operation.progress).toBe(42);

    realtime.emitOpen();
    await vi.waitFor(() => expect(getCopyJob).toHaveBeenCalledTimes(2));
  });

  it('does not restore a copied job belonging to a prior signed-in user', async () => {
    sessionStorage.setItem('facility-copy-job', JSON.stringify({ jobId, ownerId: 'user-a' }));
    const operation = new CopyOperation();

    operation.initialize('user-b');

    expect(operation.isPending).toBe(false);
    expect(getCopyJob).not.toHaveBeenCalled();
    expect(sessionStorage.getItem('facility-copy-job')).toBeNull();
  });
});
