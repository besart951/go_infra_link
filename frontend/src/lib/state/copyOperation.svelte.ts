import { ApiException } from '$lib/api/client.js';
import type { CopyJob, FacilityJob } from '$lib/domain/facility/copy-job.js';
import { copyJobRepository } from '$lib/infrastructure/api/copyJobRepository.js';
import {
  facilityReferenceDataCache,
  type FacilityCopyJobProgressEvent
} from '$lib/services/facilityReferenceDataCache.js';

const storageKey = 'facility-copy-job';

type CopyOperationResult = { started: true } | { started: false };

interface CopyOperationCallbacks {
  start: (operationId: string) => Promise<FacilityJobInput>;
  onCompleted: () => Promise<void> | void;
  onFailed: () => Promise<void> | void;
}

type FacilityJobInput = Pick<FacilityJob, 'jobId' | 'kind' | 'status' | 'progress' | 'stage'> &
  Partial<Omit<FacilityJob, 'jobId' | 'kind' | 'status' | 'progress' | 'stage'>>;

interface PersistedCopyJob {
  jobId: string;
  ownerId?: string;
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
  private unsubscribeProgress: (() => void) | null = null;
  private unsubscribeRealtimeOpen: (() => void) | null = null;
  private settlePromise: Promise<void> | null = null;
  private onOnline: (() => void) | null = null;
  private onOffline: (() => void) | null = null;
  private onFocus: (() => void) | null = null;
  private pollTimer: ReturnType<typeof setTimeout> | null = null;

  initialize(ownerId?: string | null): void {
    if (this.initialized) return;
    this.initialized = true;
    this.ownerId = ownerId ?? null;
    this.unsubscribeProgress = facilityReferenceDataCache.subscribeCopyJobProgress((event) =>
      this.handleProgress(event)
    );
    this.unsubscribeRealtimeOpen = facilityReferenceDataCache.subscribeRealtimeOpen(() => {
      void this.reconcile();
    });

    if (typeof window !== 'undefined') {
      this.onOnline = () => {
        this.connectionInterrupted = false;
        void this.reconcile();
      };
      this.onOffline = () => {
        if (this.hasActiveJobs) this.connectionInterrupted = true;
      };
      this.onFocus = () => void this.loadJobsAndRecover();
      window.addEventListener('online', this.onOnline);
      window.addEventListener('offline', this.onOffline);
      window.addEventListener('focus', this.onFocus);
    }

    const persisted = readPersistedCopyJob();
    if (!persisted) {
      void this.loadJobsAndRecover();
      return;
    }
    if (persisted.ownerId && this.ownerId && persisted.ownerId !== this.ownerId) {
      clearPersistedCopyJob();
      return;
    }

    this.jobId = persisted.jobId;
    this.isPending = true;
    this.stage = 'preparing';
    this.connectionInterrupted = typeof navigator !== 'undefined' && !navigator.onLine;
    void this.reconcile();
    this.schedulePolling();
  }

  dispose(): void {
    this.unsubscribeProgress?.();
    this.unsubscribeRealtimeOpen?.();
    this.unsubscribeProgress = null;
    this.unsubscribeRealtimeOpen = null;
    if (typeof window !== 'undefined') {
      if (this.onOnline) window.removeEventListener('online', this.onOnline);
      if (this.onOffline) window.removeEventListener('offline', this.onOffline);
      if (this.onFocus) window.removeEventListener('focus', this.onFocus);
    }
    this.onOnline = null;
    this.onOffline = null;
    this.onFocus = null;
    this.stopPolling();
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
    const normalized = normalizeJob(job);
    this.jobs = [
      normalized,
      ...this.jobs.filter((candidate) => candidate.jobId !== normalized.jobId)
    ];
    this.schedulePolling();
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
    persistCopyJob({ jobId: operationId, ...(this.ownerId ? { ownerId: this.ownerId } : {}) });

    try {
      const job = await callbacks.start(operationId);
      if (job.jobId !== this.jobId) {
        this.jobId = job.jobId;
        persistCopyJob({ jobId: job.jobId, ...(this.ownerId ? { ownerId: this.ownerId } : {}) });
      }
      this.applyJob(job);
      this.schedulePolling();
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
    const next = normalizeJob({
      ...existing,
      jobId: event.job_id,
      kind: event.kind,
      status: event.status,
      progress: event.progress,
      stage: event.stage,
      ...(event.job_type ? { type: event.job_type } : {}),
      ...(event.class ? { class: event.class } : {}),
      ...(event.processed !== undefined ? { processed: event.processed } : {}),
      ...(event.total !== undefined ? { total: event.total } : {}),
      ...(event.success_count !== undefined ? { successCount: event.success_count } : {}),
      ...(event.failure_count !== undefined ? { failureCount: event.failure_count } : {}),
      ...(event.error ? { error: event.error } : {})
    });
    this.jobs = [next, ...this.jobs.filter((candidate) => candidate.jobId !== next.jobId)];
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
      const jobs = (await copyJobRepository.list()).map(normalizeJob);
      this.jobs = jobs;
      const active = jobs.find(
        (job) => (job.status === 'queued' || job.status === 'running') && job.type !== 'export'
      );
      if (!active) return;

      this.jobId = active.jobId;
      this.isPending = true;
      persistCopyJob({ jobId: active.jobId, ...(this.ownerId ? { ownerId: this.ownerId } : {}) });
      this.applyJob(active);
      this.schedulePolling();
    } catch {
      this.connectionInterrupted = true;
    }
  }

  private applyJob(input: FacilityJobInput): void {
    const job = normalizeJob(input);
    if (job.jobId !== this.jobId) return;

    this.jobs = [job, ...this.jobs.filter((candidate) => candidate.jobId !== job.jobId)];

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
    this.stopPolling();
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
    this.callbacks = null;
    this.settlePromise = null;
    clearPersistedCopyJob();
  }

  private schedulePolling(): void {
    if (import.meta.env.MODE === 'test' || !this.hasActiveJobs || this.pollTimer) return;
    this.pollTimer = setTimeout(() => {
      this.pollTimer = null;
      void this.loadJobsAndRecover().finally(() => this.schedulePolling());
    }, 5_000);
  }

  private stopPolling(): void {
    if (!this.pollTimer) return;
    clearTimeout(this.pollTimer);
    this.pollTimer = null;
  }
}

function normalizeJob(job: FacilityJobInput): CopyJob {
  return {
    ...job,
    type: job.type ?? 'copy',
    class: job.class ?? (job.type === 'export' ? 'export' : 'mutation'),
    attempts: job.attempts ?? 0,
    processed: job.processed ?? 0,
    successCount: job.successCount ?? 0,
    failureCount: job.failureCount ?? 0,
    retryable: job.retryable ?? false
  };
}

function isNetworkInterruption(error: unknown): boolean {
  return error instanceof ApiException && error.status === 0;
}

function readPersistedCopyJob(): PersistedCopyJob | null {
  if (typeof sessionStorage === 'undefined') return null;
  try {
    const parsed: unknown = JSON.parse(sessionStorage.getItem(storageKey) ?? 'null');
    if (
      !parsed ||
      typeof parsed !== 'object' ||
      !('jobId' in parsed) ||
      typeof parsed.jobId !== 'string'
    ) {
      return null;
    }
    return {
      jobId: parsed.jobId,
      ...('ownerId' in parsed && typeof parsed.ownerId === 'string'
        ? { ownerId: parsed.ownerId }
        : {})
    };
  } catch {
    return null;
  }
}

function persistCopyJob(job: PersistedCopyJob): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.setItem(storageKey, JSON.stringify(job));
}

function clearPersistedCopyJob(): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.removeItem(storageKey);
}

export { FacilityJobState as CopyOperation };

export const facilityJobState = new FacilityJobState();
export const copyOperation = facilityJobState;
