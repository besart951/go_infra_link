import { writable } from 'svelte/store';

export interface NetworkStatus {
  browserOnline: boolean;
  apiUnavailable: boolean;
  retrying: boolean;
  retryAttempt: number;
  retryMax: number;
  lastFailureAt: number | null;
  lastSuccessAt: number | null;
}

const initialStatus: NetworkStatus = {
  browserOnline: true,
  apiUnavailable: false,
  retrying: false,
  retryAttempt: 0,
  retryMax: 0,
  lastFailureAt: null,
  lastSuccessAt: null
};

let initialized = false;

export const networkStatus = writable<NetworkStatus>(initialStatus);

export function initNetworkStatus(): void {
  if (initialized || typeof window === 'undefined') return;
  initialized = true;

  const updateBrowserStatus = () => {
    const browserOnline = navigator.onLine !== false;
    networkStatus.update((state) => ({
      ...state,
      browserOnline,
      apiUnavailable: browserOnline ? state.apiUnavailable : true,
      retrying: browserOnline ? state.retrying : false
    }));
  };

  updateBrowserStatus();
  window.addEventListener('online', () => {
    networkStatus.update((state) => ({
      ...state,
      browserOnline: true,
      apiUnavailable: false,
      retrying: false
    }));
  });
  window.addEventListener('offline', updateBrowserStatus);
}

export function reportApiRetry(attempt: number, maxRetries: number): void {
  networkStatus.update((state) => ({
    ...state,
    apiUnavailable: true,
    retrying: true,
    retryAttempt: attempt,
    retryMax: maxRetries,
    lastFailureAt: Date.now()
  }));
}

export function reportApiFailure(): void {
  networkStatus.update((state) => ({
    ...state,
    apiUnavailable: true,
    retrying: false,
    retryAttempt: 0,
    retryMax: 0,
    lastFailureAt: Date.now()
  }));
}

export function reportApiSuccess(): void {
  networkStatus.update((state) => ({
    ...state,
    apiUnavailable: false,
    retrying: false,
    retryAttempt: 0,
    retryMax: 0,
    lastSuccessAt: Date.now()
  }));
}
