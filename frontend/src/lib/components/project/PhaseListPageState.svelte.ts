import { deletePhase } from '$lib/infrastructure/api/phase.adapter.js';
import { addToast } from '$lib/components/toast.svelte';
import type { Phase } from '$lib/domain/phase/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { phaseListStore } from '$lib/stores/phases/phaseListStore.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

export class PhaseListPageState {
  showForm = $state(false);
  editingPhase: Phase | undefined = $state(undefined);
  deleting = $state(false);

  initialize(): void {
    phaseListStore.load();
  }

  handleEdit(phase: Phase): void {
    this.editingPhase = phase;
    this.showForm = true;
  }

  handleCreate(): void {
    this.editingPhase = undefined;
    this.showForm = true;
  }

  handleSuccess(): void {
    this.showForm = false;
    this.editingPhase = undefined;
    phaseListStore.reload();
  }

  handleCancel(): void {
    this.showForm = false;
    this.editingPhase = undefined;
  }

  async handleDelete(phase: Phase): Promise<void> {
    const ok = await confirm({
      title: translate('phases.confirm.delete_title'),
      message: translate('phases.confirm.delete_message', { name: phase.name }),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!ok) return;
    this.deleting = true;
    try {
      await deletePhase(phase.id);
      addToast(translate('phases.toasts.deleted'), 'success');
      phaseListStore.reload();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('phases.toasts.delete_failed'),
        'error'
      );
    } finally {
      this.deleting = false;
    }
  }
}
