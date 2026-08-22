import {
  facilityReferenceDataCache,
  type FacilityCopyJobProgressEvent
} from '$lib/services/facilityReferenceDataCache.js';

export interface FacilityJobRecoveryActions {
  hasActiveJobs: () => boolean;
  onProgress: (event: FacilityCopyJobProgressEvent) => void;
  reconcile: () => Promise<void>;
  refresh: () => Promise<void>;
  setInterrupted: (interrupted: boolean) => void;
}

export class FacilityJobRecovery {
  private pollTimer: ReturnType<typeof setTimeout> | null = null;
  private unsubscribers: (() => void)[] = [];

  constructor(private readonly actions: FacilityJobRecoveryActions) {}

  start(): void {
    this.unsubscribers = [
      facilityReferenceDataCache.subscribeCopyJobProgress(this.actions.onProgress),
      facilityReferenceDataCache.subscribeRealtimeOpen(() => void this.actions.reconcile())
    ];
    if (typeof window === 'undefined') return;
    window.addEventListener('online', this.onOnline);
    window.addEventListener('offline', this.onOffline);
    window.addEventListener('focus', this.onFocus);
  }

  stop(): void {
    this.unsubscribers.forEach((unsubscribe) => unsubscribe());
    this.unsubscribers = [];
    if (typeof window !== 'undefined') {
      window.removeEventListener('online', this.onOnline);
      window.removeEventListener('offline', this.onOffline);
      window.removeEventListener('focus', this.onFocus);
    }
    this.pausePolling();
  }

  schedule(): void {
    if (import.meta.env.MODE === 'test' || !this.actions.hasActiveJobs() || this.pollTimer) return;
    this.pollTimer = setTimeout(() => {
      this.pollTimer = null;
      void this.actions.refresh().finally(() => this.schedule());
    }, 5_000);
  }

  pausePolling(): void {
    if (!this.pollTimer) return;
    clearTimeout(this.pollTimer);
    this.pollTimer = null;
  }

  private onOnline = (): void => {
    this.actions.setInterrupted(false);
    void this.actions.reconcile();
  };

  private onOffline = (): void => {
    if (this.actions.hasActiveJobs()) this.actions.setInterrupted(true);
  };

  private onFocus = (): void => void this.actions.refresh();
}
