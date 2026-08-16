import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

export const toasts = writable<Toast[]>([]);

export function addToast(message: string, type: ToastType = 'info', duration = 5000): string {
  const id = Math.random().toString(36).substring(2, 9);
  toasts.update((all) => [...all, { id, message, type }]);

  if (duration > 0) {
    setTimeout(() => {
      removeToast(id);
    }, duration);
  }

  return id;
}

export function removeToast(id: string): void {
  toasts.update((all) => all.filter((toast) => toast.id !== id));
}
