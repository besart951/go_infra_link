import { t } from '$lib/i18n/index.js';
import type {
  BaseDataTableStateOptions,
  DataTableFetchStrategy,
  DataTableQuery,
  SortOrder,
  TableFilterRecord
} from './contracts.js';
import { sanitizeFilters } from './sanitizeFilters.js';

export abstract class BaseDataTableState<
  TItem extends { id: string },
  TFilters extends TableFilterRecord
> {
  readonly pageSize: number;

  items = $state<TItem[]>([]);
  total = $state(0);
  page = $state(1);
  totalPages = $state(0);
  searchText = $state('');
  orderBy = $state<string | undefined>(undefined);
  order = $state<SortOrder | undefined>(undefined);
  filters = $state<TFilters>({} as TFilters);
  loading = $state(false);
  error = $state<string | null>(null);
  selectedIds = $state<Set<string>>(new Set());

  readonly selectedCount = $derived.by(() => this.selectedIds.size);
  readonly allSelected = $derived.by(
    () => this.items.length > 0 && this.items.every((item) => this.selectedIds.has(item.id))
  );
  readonly someSelected = $derived.by(
    () => this.items.some((item) => this.selectedIds.has(item.id)) && !this.allSelected
  );
  readonly hasActiveFilters = $derived.by(() =>
    Object.values(this.filters).some((value) => value !== undefined && value !== '')
  );

  private abortController: AbortController | null = null;
  private readonly initialFilters: TFilters;

  constructor(
    protected readonly strategy: DataTableFetchStrategy<TItem, TFilters>,
    options: BaseDataTableStateOptions<TFilters> = {}
  ) {
    this.pageSize = options.pageSize ?? 50;
    this.initialFilters = sanitizeFilters(options.initialFilters ?? ({} as TFilters));
    this.filters = { ...this.initialFilters };
  }

  protected onSelectionChanged() {}

  protected createQuery(): DataTableQuery<TFilters> {
    return {
      page: this.page,
      pageSize: this.pageSize,
      searchText: this.searchText,
      orderBy: this.orderBy,
      order: this.order,
      filters: sanitizeFilters({ ...this.filters })
    };
  }

  async load(): Promise<void> {
    this.abortController?.abort();

    const controller = new AbortController();
    this.abortController = controller;
    this.loading = true;
    this.error = null;

    try {
      const response = await this.strategy.fetch(this.createQuery(), controller.signal);
      if (this.abortController !== controller) return;

      this.items = response.items;
      this.total = response.total;
      this.page = response.page;
      this.totalPages = response.totalPages;
      this.error = null;
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') {
        return;
      }
      if (this.abortController !== controller) return;

      this.error = error instanceof Error ? error.message : t('facility.fetch_failed');
    } finally {
      if (this.abortController === controller) {
        this.loading = false;
      }
    }
  }

  async reload(): Promise<void> {
    await this.load();
  }

  async search(text: string): Promise<void> {
    this.searchText = text;
    this.page = 1;
    await this.load();
  }

  async setFilters(filters: TFilters): Promise<void> {
    this.filters = sanitizeFilters(filters);
    this.page = 1;
    await this.load();
  }

  async clearFilter(filterKey: keyof TFilters): Promise<void> {
    const nextFilters = { ...this.filters };
    delete nextFilters[filterKey];
    this.filters = sanitizeFilters(nextFilters as TFilters);
    this.page = 1;
    await this.load();
  }

  async clearAllFilters(): Promise<void> {
    this.filters = { ...this.initialFilters };
    this.page = 1;
    await this.load();
  }

  async setSort(orderBy?: string, order?: SortOrder): Promise<void> {
    this.orderBy = orderBy;
    this.order = order;
    this.page = 1;
    await this.load();
  }

  async toggleSort(orderBy: string): Promise<void> {
    if (this.orderBy !== orderBy) {
      await this.setSort(orderBy, 'asc');
      return;
    }

    if (this.order === 'asc') {
      await this.setSort(orderBy, 'desc');
      return;
    }

    await this.setSort(undefined, undefined);
  }

  sortState(orderBy: string): SortOrder | undefined {
    if (this.orderBy !== orderBy) return undefined;
    return this.order === 'desc' ? 'desc' : 'asc';
  }

  async goToPage(page: number): Promise<void> {
    this.page = Math.max(1, page);
    await this.load();
  }

  async goToPreviousPage(): Promise<void> {
    if (this.page <= 1) return;
    await this.goToPage(this.page - 1);
  }

  async goToNextPage(): Promise<void> {
    if (this.page >= this.totalPages) return;
    await this.goToPage(this.page + 1);
  }

  isSelected(id: string): boolean {
    return this.selectedIds.has(id);
  }

  toggleSelection(id: string): void {
    const nextSelectedIds = new Set(this.selectedIds);
    if (nextSelectedIds.has(id)) {
      nextSelectedIds.delete(id);
    } else {
      nextSelectedIds.add(id);
    }

    this.selectedIds = nextSelectedIds;
    this.onSelectionChanged();
  }

  toggleSelectAll(): void {
    this.selectedIds = this.allSelected ? new Set() : new Set(this.items.map((item) => item.id));
    this.onSelectionChanged();
  }

  clearSelection(): void {
    if (this.selectedIds.size === 0) return;
    this.selectedIds = new Set();
    this.onSelectionChanged();
  }

  replaceItem(updated: TItem) {
    this.items = this.items.map((item) => (item.id === updated.id ? updated : item));
  }

  replaceItems(updatedItems: TItem[]) {
    const itemsById = new Map(updatedItems.map((item) => [item.id, item]));
    this.items = this.items.map((item) => itemsById.get(item.id) ?? item);
  }

  dispose() {
    this.abortController?.abort();
  }
}
