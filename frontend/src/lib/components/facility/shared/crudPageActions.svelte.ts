interface ConfirmOptions {
  title: string;
  message: string;
  confirmText: string;
  cancelText: string;
  variant?: 'default' | 'destructive';
}

export interface CrudPageActionsOptions<TItem> {
  reload: () => void | Promise<void>;
  deleteItem: (item: TItem) => Promise<void>;
  confirmDelete: (options: ConfirmOptions) => Promise<boolean>;
  confirmImpactUpdate?: (options: ConfirmOptions) => Promise<boolean>;
  addToast: (message: string, type: 'success' | 'error') => void;
  getItemId?: (item: TItem) => string;
  getDeleteMessage: (item: TItem) => string;
  getDeleteSuccessMessage: () => string;
  getDeleteFailureMessage: () => string;
  getDeleteErrorMessage?: (error: unknown) => string;
  getDeleteTitle: () => string;
  getDeleteConfirmText: () => string;
  getDeleteCancelText: () => string;
  getBacnetUsageMessage?: (count: number) => string;
  getBacnetDeleteBlockedMessage?: (count: number) => string;
  getBacnetUpdateConfirmTitle?: () => string;
  getBacnetUpdateConfirmMessage?: (count: number) => string;
  getBacnetUpdateConfirmAgainTitle?: () => string;
  getBacnetUpdateConfirmAgainMessage?: (count: number) => string;
  getBacnetUpdateConfirmText?: () => string;
}

export class CrudPageActions<TItem> {
  showForm = $state(false);
  editingItem = $state<TItem | undefined>(undefined);
  bacnetUsageCounts = $state<Record<string, number>>({});

  constructor(private readonly options: CrudPageActionsOptions<TItem>) {}

  create(): void {
    this.editingItem = undefined;
    this.showForm = true;
  }

  edit(item: TItem): void {
    this.editingItem = item;
    this.showForm = true;
  }

  async success(): Promise<void> {
    this.showForm = false;
    this.editingItem = undefined;
    await this.options.reload();
  }

  cancel(): void {
    this.showForm = false;
    this.editingItem = undefined;
  }

  async copy(value: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  setBacnetUsageCounts(counts: Record<string, number>): void {
    this.bacnetUsageCounts = { ...counts };
  }

  mergeBacnetUsageCounts(counts: Record<string, number>): void {
    this.bacnetUsageCounts = { ...this.bacnetUsageCounts, ...counts };
  }

  getBacnetUsageCount(item: TItem): number {
    const id = this.getItemId(item);
    return id ? (this.bacnetUsageCounts[id] ?? 0) : 0;
  }

  isBacnetLinked(item: TItem): boolean {
    return this.getBacnetUsageCount(item) > 0;
  }

  isDeleteDisabled(item: TItem): boolean {
    return this.isBacnetLinked(item);
  }

  getBacnetUsageMessage(item: TItem): string {
    const count = this.getBacnetUsageCount(item);
    if (count <= 0) return '';
    return this.options.getBacnetUsageMessage?.(count) ?? '';
  }

  getDeleteLabel(item: TItem, fallback: string): string {
    const count = this.getBacnetUsageCount(item);
    if (count <= 0) return fallback;
    return this.options.getBacnetDeleteBlockedMessage?.(count) ?? fallback;
  }

  async confirmBacnetImpactUpdate(item: TItem): Promise<boolean> {
    const count = this.getBacnetUsageCount(item);
    if (count <= 0 || !this.options.confirmImpactUpdate) {
      return true;
    }

    const first = await this.options.confirmImpactUpdate({
      title: this.options.getBacnetUpdateConfirmTitle?.() ?? '',
      message: this.options.getBacnetUpdateConfirmMessage?.(count) ?? '',
      confirmText:
        this.options.getBacnetUpdateConfirmText?.() ?? this.options.getDeleteConfirmText(),
      cancelText: this.options.getDeleteCancelText()
    });
    if (!first) return false;

    return this.options.confirmImpactUpdate({
      title: this.options.getBacnetUpdateConfirmAgainTitle?.() ?? '',
      message: this.options.getBacnetUpdateConfirmAgainMessage?.(count) ?? '',
      confirmText:
        this.options.getBacnetUpdateConfirmText?.() ?? this.options.getDeleteConfirmText(),
      cancelText: this.options.getDeleteCancelText()
    });
  }

  async delete(item: TItem): Promise<void> {
    if (this.isDeleteDisabled(item)) {
      const count = this.getBacnetUsageCount(item);
      this.options.addToast(
        this.options.getBacnetDeleteBlockedMessage?.(count) ??
          this.options.getDeleteFailureMessage(),
        'error'
      );
      return;
    }

    const ok = await this.options.confirmDelete({
      title: this.options.getDeleteTitle(),
      message: this.options.getDeleteMessage(item),
      confirmText: this.options.getDeleteConfirmText(),
      cancelText: this.options.getDeleteCancelText(),
      variant: 'destructive'
    });
    if (!ok) return;

    try {
      await this.options.deleteItem(item);
      this.options.addToast(this.options.getDeleteSuccessMessage(), 'success');
      await this.options.reload();
    } catch (error) {
      this.options.addToast(
        this.options.getDeleteErrorMessage?.(error) ??
          (error instanceof Error ? error.message : this.options.getDeleteFailureMessage()),
        'error'
      );
    }
  }

  private getItemId(item: TItem): string {
    if (this.options.getItemId) return this.options.getItemId(item);
    if (typeof item === 'string') return item;
    const id = (item as { id?: unknown })?.id;
    return typeof id === 'string' ? id : '';
  }
}
