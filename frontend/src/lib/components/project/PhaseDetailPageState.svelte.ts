import { goto } from '$app/navigation';
import { deletePhase, getPhase } from '$lib/infrastructure/api/phase.adapter.js';
import { addToast } from '$lib/components/toast.svelte';
import type { Phase } from '$lib/domain/phase/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

export class PhaseDetailPageState {
  phase = $state<Phase | null>(null);
  loading = $state(true);
  error = $state<string | null>(null);
  busy = $state(false);

  constructor(private readonly resolvePhaseId: () => string) {}

  get phaseId(): string {
    return this.resolvePhaseId();
  }

  async load(): Promise<void> {
    if (!this.phaseId) {
      this.error = translate('phases.errors.missing_id');
      this.loading = false;
      return;
    }

    this.loading = true;
    this.error = null;
    try {
      this.phase = await getPhase(this.phaseId);
    } catch (error) {
      this.error = error instanceof Error ? error.message : translate('phases.errors.load_failed');
    } finally {
      this.loading = false;
    }
  }

  async handleDelete(): Promise<void> {
    if (!this.phase) return;
    const ok = await confirm({
      title: translate('phases.confirm.delete_title'),
      message: translate('phases.confirm.delete_message', { name: this.phase.name }),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!ok) return;
    this.busy = true;
    try {
      await deletePhase(this.phase.id);
      addToast(translate('phases.toasts.deleted'), 'success');
      await goto('/projects/phases');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('phases.toasts.delete_failed'),
        'error'
      );
    } finally {
      this.busy = false;
    }
  }
}
