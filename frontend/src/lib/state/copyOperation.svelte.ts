import { ApiException } from '$lib/api/client.js';
import type { CopyJob } from '$lib/domain/facility/copy-job.js';
import { copyJobRepository } from '$lib/infrastructure/api/copyJobRepository.js';
import type { FacilityCopyJobProgressEvent } from '$lib/services/facilityReferenceDataCache.js';
import {
  normalizeFacilityJob,
  reconcileFacilityJob,
  sameFacilityJobProgress,
  type FacilityJobInput
} from './facilityJobModel.js';
import {
  clearFacilityJobHint,
  persistFacilityJobHint,
  readFacilityJobHint
} from './facilityJobSession.js';
import { FacilityJobRecovery } from './facilityJobRecovery.js';

type CopyOperationResult = { started: true } | { started: false };

interface CopyOperationCallbacks {
  start: (operationId: string) => Promise<FacilityJobInput>;
  onCompleted: () => Promise<void> | void;
  onFailed: () => Promise<void> | void;
}

/**
 * Tracks the one active hierarchy copy for this browser tab. The server owns
 * the actual job and pushes its progress via the already-open facility stream.
 * The job ID is persisted so a lost internet connection can be reconciled on
 * reconnect without enabling a second copy action.
 */
export class FacilityJobState {
  isPending = $state(false);
  progress = $state(0);
  stage = $state('queued');
  connectionInterrupted = $state(false);
  jobs = $state.raw<CopyJob[]>([]);

  private initialized = false;
  private jobId: string | null = null;
  private ownerId: string | null = null;
  private callbacks: CopyOperationCallbacks | null = null;
  private settlePromise: Promise<void> | null = null;
  private recovery: FacilityJobRecovery | null = null;

  initialize(ownerId?: string | null): void {
    if (this.initialized) return;
    this.initialized = true;
    this.ownerId = ownerId ?? null;
    this.recovery = new FacilityJobRecovery({
      hasActiveJobs: () => this.hasActiveJobs,
      onProgress: (event) => this.handleProgress(event),
      reconcile: () => this.reconcile(),
      refresh: () => this.loadJobsAndRecover(),
      setInterrupted: (interrupted) => {
        this.connectionInterrupted = interrupted;
      }
    });
    this.recovery.start();

    const persisted = readFacilityJobHint();
    if (!persisted) {
      void this.loadJobsAndRecover();
      return;
    }
    if (persisted.ownerId && this.ownerId && persisted.ownerId !== this.ownerId) {
      clearFacilityJobHint();
      return;
    }

    this.jobId = persisted.jobId;
    this.isPending = true;
    this.stage = 'preparing';
    this.connectionInterrupted = typeof navigator !== 'undefined' && !navigator.onLine;
    void this.reconcile();
    this.recovery.schedule();
  }

  dispose(): void {
    this.recovery?.stop();
    this.recovery = null;
    this.initialized = false;
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
    this.callbacks = null;
    this.settlePromise = null;
  }

  get activeJobs(): CopyJob[] {
    return this.jobs.filter((job) => job.status === 'queued' || job.status === 'running');
  }

  get exportJobs(): CopyJob[] {
    return this.jobs.filter((job) => job.type === 'export');
  }

  get hasActiveJobs(): boolean {
    return this.activeJobs.length > 0 || this.isPending;
  }

  track(job: FacilityJobInput): void {
    this.jobs = reconcileFacilityJob(this.jobs, job);
    this.recovery?.schedule();
  }

  async retry(jobId: string): Promise<void> {
    const job = await copyJobRepository.retry(jobId);
    this.track(job);
  }

  async run(callbacks: CopyOperationCallbacks): Promise<CopyOperationResult> {
    this.initialize();
    if (this.isPending) return { started: false };

    const operationId = crypto.randomUUID();
    this.jobId = operationId;
    this.callbacks = callbacks;
    this.isPending = true;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    persistFacilityJobHint({
      jobId: operationId,
      ...(this.ownerId ? { ownerId: this.ownerId } : {})
    });

    try {
      const job = await callbacks.start(operationId);
      if (job.jobId !== this.jobId) {
        this.jobId = job.jobId;
        persistFacilityJobHint({
          jobId: job.jobId,
          ...(this.ownerId ? { ownerId: this.ownerId } : {})
        });
      }
      this.applyJob(job);
      this.recovery?.schedule();
      return { started: true };
    } catch (error) {
      if (isNetworkInterruption(error)) {
        this.connectionInterrupted = true;
        return { started: true };
      }
      this.reset();
      throw error;
    }
  }

  private handleProgress(event: FacilityCopyJobProgressEvent): void {
    const existing = this.jobs.find((job) => job.jobId === event.job_id);
    const next = normalizeFacilityJob({
      ...existing,
      jobId: event.job_id,
      kind: event.kind,
      status: event.status,
      progress: event.progress,
      stage: event.stage,
      updatedAt: event.updated_at,
      ...(event.job_type ? { type: event.job_type } : {}),
      ...(event.class ? { class: event.class } : {}),
      ...(event.processed !== undefined ? { processed: event.processed } : {}),
      ...(event.total !== undefined ? { total: event.total } : {}),
      ...(event.success_count !== undefined ? { successCount: event.success_count } : {}),
      ...(event.failure_count !== undefined ? { failureCount: event.failure_count } : {}),
      ...(event.error ? { error: event.error } : {})
    });
    if (existing && sameFacilityJobProgress(existing, next)) return;
    this.jobs = reconcileFacilityJob(this.jobs, next);
    if (event.job_id === this.jobId) this.applyJob(next);
    if (event.status === 'completed' || event.status === 'failed') {
      void this.refreshJob(event.job_id);
    }
  }

  private async refreshJob(jobId: string): Promise<void> {
    try {
      this.track(await copyJobRepository.get(jobId));
    } catch {
      this.connectionInterrupted = true;
    }
  }

  private async reconcile(): Promise<void> {
    if (!this.isPending || !this.jobId) return;

    try {
      this.connectionInterrupted = false;
      this.applyJob(await copyJobRepository.get(this.jobId));
    } catch (error) {
      if (error instanceof ApiException && error.status === 404) {
        await this.fail();
        return;
      }
      this.connectionInterrupted = true;
    }
  }

  private async loadJobsAndRecover(): Promise<void> {
    try {
      const jobs = (await copyJobRepository.list()).map(normalizeFacilityJob);
      this.jobs = jobs;
      const active = jobs.find(
        (job) => (job.status === 'queued' || job.status === 'running') && job.type !== 'export'
      );
      if (!active) return;

      this.jobId = active.jobId;
      this.isPending = true;
      persistFacilityJobHint({
        jobId: active.jobId,
        ...(this.ownerId ? { ownerId: this.ownerId } : {})
      });
      this.applyJob(active);
      this.recovery?.schedule();
    } catch {
      this.connectionInterrupted = true;
    }
  }

  private applyJob(input: FacilityJobInput): void {
    const job = normalizeFacilityJob(input);
    if (job.jobId !== this.jobId) return;

    this.jobs = reconcileFacilityJob(this.jobs, job);

    this.progress = job.progress;
    this.stage = job.stage;
    this.connectionInterrupted = false;
    if (job.status === 'completed') {
      void this.complete();
    } else if (job.status === 'failed') {
      void this.fail();
    }
  }

  private async complete(): Promise<void> {
    if (this.settlePromise) return this.settlePromise;
    this.settlePromise = (async () => {
      try {
        await this.callbacks?.onCompleted();
      } finally {
        this.reset();
      }
    })();
    return this.settlePromise;
  }

  private async fail(): Promise<void> {
    if (this.settlePromise) return this.settlePromise;
    this.settlePromise = (async () => {
      try {
        await this.callbacks?.onFailed();
      } finally {
        this.reset();
      }
    })();
    return this.settlePromise;
  }

  private reset(): void {
    this.recovery?.pausePolling();
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
    this.callbacks = null;
    this.settlePromise = null;
    clearFacilityJobHint();
  }
}

function isNetworkInterruption(error: unknown): boolean {
  return error instanceof ApiException && error.status === 0;
}

export { FacilityJobState as CopyOperation };

export const facilityJobState = new FacilityJobState();
export const copyOperation = facilityJobState;
