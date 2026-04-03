import { writable } from 'svelte/store';

export interface ConfirmDialogOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'default' | 'destructive';
}

export interface ConfirmDialogState extends ConfirmDialogOptions {
  open: boolean;
  onConfirm?: () => void;
  onCancel?: () => void;
}

export const confirmDialogState = writable<ConfirmDialogState>({
  open: false,
  title: '',
  message: '',
  confirmText: 'Bestätigen',
  cancelText: 'Abbrechen',
  variant: 'default'
});

function closeDialog() {
  confirmDialogState.update((state) => ({ ...state, open: false }));
}

export function confirm(options: ConfirmDialogOptions): Promise<boolean> {
  return new Promise((resolve) => {
    confirmDialogState.set({
      ...options,
      open: true,
      confirmText: options.confirmText || 'Bestätigen',
      cancelText: options.cancelText || 'Abbrechen',
      variant: options.variant || 'default',
      onConfirm: () => {
        closeDialog();
        resolve(true);
      },
      onCancel: () => {
        closeDialog();
        resolve(false);
      }
    });
  });
}
