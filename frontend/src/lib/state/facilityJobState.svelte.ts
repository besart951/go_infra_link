import { ApiException } from '$lib/api/client.js';
import type { FacilityJob } from '$lib/domain/facility/facility-job.js';
import { facilityJobRepository } from '$lib/infrastructure/api/facilityJobRepository.js';
import type { FacilityJobProgressEvent } from '$lib/services/facilityReferenceDataCache.js';
import {
  normalizeFacilityJob,
  reconcileFacilityJob,
  sameFacilityJobProgress,
  type FacilityJobInput
} from './facilityJobModel.js';
import { FacilityJobRecovery } from './facilityJobRecovery.js';

type FacilityJobSubmissionResult = { started: true } | { started: false };

type FacilityJobStarter = (operationId: string) => Promise<FacilityJobInput>;

/**
 * Tracks persisted Facility jobs. PostgreSQL-backed job responses remain the
 * source of truth after reloads and reconnects.
 */
export class FacilityJobState {
  isPending = $state(false);
  progress = $state(0);
  stage = $state('queued');
  connectionInterrupted = $state(false);
  jobs = $state.raw<FacilityJob[]>([]);

  private initialized = false;
  private jobId: string | null = null;
  private ownerId: string | null = null;
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

    this.connectionInterrupted = typeof navigator !== 'undefined' && !navigator.onLine;
    void this.loadJobsAndRecover();
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
  }

  get activeJobs(): FacilityJob[] {
    return this.jobs.filter((job) => job.status === 'queued' || job.status === 'running');
  }

  get exportJobs(): FacilityJob[] {
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
    const job = await facilityJobRepository.retry(jobId);
    this.track(job);
  }

  async submit(start: FacilityJobStarter): Promise<FacilityJobSubmissionResult> {
    this.initialize();
    if (this.isPending) return { started: false };

    const operationId = crypto.randomUUID();
    this.jobId = operationId;
    this.isPending = true;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    try {
      const job = await start(operationId);
      if (job.jobId !== this.jobId) {
        this.jobId = job.jobId;
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

  private handleProgress(event: FacilityJobProgressEvent): void {
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
      this.track(await facilityJobRepository.get(jobId));
    } catch {
      this.connectionInterrupted = true;
    }
  }

  private async reconcile(): Promise<void> {
    if (!this.isPending || !this.jobId) return;

    try {
      this.connectionInterrupted = false;
      this.applyJob(await facilityJobRepository.get(this.jobId));
    } catch (error) {
      if (error instanceof ApiException && error.status === 404) {
        this.reset();
        return;
      }
      this.connectionInterrupted = true;
    }
  }

  private async loadJobsAndRecover(): Promise<void> {
    try {
      const page = await facilityJobRepository.list();
      const jobs = page.items.map(normalizeFacilityJob);
      this.jobs = jobs;
      const active = jobs.find(
        (job) => (job.status === 'queued' || job.status === 'running') && job.type !== 'export'
      );
      if (!active) return;

      this.jobId = active.jobId;
      this.isPending = true;
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
      this.reset();
    } else if (job.status === 'failed') {
      this.reset();
    }
  }

  private reset(): void {
    this.recovery?.pausePolling();
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
  }
}

function isNetworkInterruption(error: unknown): boolean {
  return error instanceof ApiException && error.status === 0;
}

export const facilityJobState = new FacilityJobState();
