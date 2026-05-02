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
  addToast: (message: string, type: 'success' | 'error') => void;
  getDeleteMessage: (item: TItem) => string;
  getDeleteSuccessMessage: () => string;
  getDeleteFailureMessage: () => string;
  getDeleteErrorMessage?: (error: unknown) => string;
  getDeleteTitle: () => string;
  getDeleteConfirmText: () => string;
  getDeleteCancelText: () => string;
}

export class CrudPageActions<TItem> {
  showForm = $state(false);
  editingItem = $state<TItem | undefined>(undefined);

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

  async delete(item: TItem): Promise<void> {
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
}
