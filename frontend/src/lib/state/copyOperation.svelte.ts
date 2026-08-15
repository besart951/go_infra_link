import { ApiException } from '$lib/api/client.js';
import type { CopyJob } from '$lib/domain/facility/copy-job.js';
import { copyJobRepository } from '$lib/infrastructure/api/copyJobRepository.js';
import {
  facilityReferenceDataCache,
  type FacilityCopyJobProgressEvent
} from '$lib/services/facilityReferenceDataCache.js';

const storageKey = 'facility-copy-job';

type CopyOperationResult = { started: true } | { started: false };

interface CopyOperationCallbacks {
  start: (operationId: string) => Promise<CopyJob>;
  onCompleted: () => Promise<void> | void;
  onFailed: () => Promise<void> | void;
}

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
export class CopyOperation {
  isPending = $state(false);
  progress = $state(0);
  stage = $state('queued');
  connectionInterrupted = $state(false);

  private initialized = false;
  private jobId: string | null = null;
  private ownerId: string | null = null;
  private callbacks: CopyOperationCallbacks | null = null;
  private unsubscribeProgress: (() => void) | null = null;
  private unsubscribeRealtimeOpen: (() => void) | null = null;
  private settlePromise: Promise<void> | null = null;
  private onOnline: (() => void) | null = null;
  private onOffline: (() => void) | null = null;

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
        if (this.isPending) this.connectionInterrupted = true;
      };
      window.addEventListener('online', this.onOnline);
      window.addEventListener('offline', this.onOffline);
    }

    const persisted = readPersistedCopyJob();
    if (!persisted) return;
    if (persisted.ownerId && this.ownerId && persisted.ownerId !== this.ownerId) {
      clearPersistedCopyJob();
      return;
    }

    this.jobId = persisted.jobId;
    this.isPending = true;
    this.stage = 'preparing';
    this.connectionInterrupted = typeof navigator !== 'undefined' && !navigator.onLine;
    void this.reconcile();
  }

  dispose(): void {
    this.unsubscribeProgress?.();
    this.unsubscribeRealtimeOpen?.();
    this.unsubscribeProgress = null;
    this.unsubscribeRealtimeOpen = null;
    if (typeof window !== 'undefined') {
      if (this.onOnline) window.removeEventListener('online', this.onOnline);
      if (this.onOffline) window.removeEventListener('offline', this.onOffline);
    }
    this.onOnline = null;
    this.onOffline = null;
    this.initialized = false;
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
    this.callbacks = null;
    this.settlePromise = null;
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
    if (!this.isPending || event.job_id !== this.jobId) return;

    this.applyJob({
      jobId: event.job_id,
      kind: event.kind,
      status: event.status,
      progress: event.progress,
      stage: event.stage,
      ...(event.error ? { error: event.error } : {})
    });
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

  private applyJob(job: CopyJob): void {
    if (job.jobId !== this.jobId) return;

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
    this.isPending = false;
    this.progress = 0;
    this.stage = 'queued';
    this.connectionInterrupted = false;
    this.jobId = null;
    this.callbacks = null;
    this.settlePromise = null;
    clearPersistedCopyJob();
  }
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

export const copyOperation = new CopyOperation();
