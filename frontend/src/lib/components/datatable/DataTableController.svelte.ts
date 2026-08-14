import {
  createDefaultDataTableQuery,
  normalizeDataTableQuery,
  normalizePage,
  normalizePageSize
} from './dataTableModel.js';
import type {
  DataTableFilterRecord,
  DataTableLoadQuery,
  DataTableQueryState,
  DataTableRowId,
  DataTableSortState,
  DataTableSource
} from './types.js';

export interface DataTableControllerOptions<
  TItem,
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  source?: DataTableSource<TItem, TFilters>;
  initialRows?: TItem[];
  initialTotalRows?: number;
  initialQuery?: Partial<DataTableQueryState<TFilters>>;
  getRowId?: (row: TItem) => DataTableRowId;
  autoLoad?: boolean;
}

export class DataTableController<
  TItem,
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  rows = $state<TItem[]>([]);
  totalRows = $state(0);
  totalPages = $state(1);
  loading = $state(false);
  error = $state<string | null>(null);
  query = $state<DataTableQueryState<TFilters>>(createDefaultDataTableQuery<TFilters>());
  selectedIds = $state<Set<DataTableRowId>>(new Set());
  expandedIds = $state<Set<DataTableRowId>>(new Set());

  private readonly source?: DataTableSource<TItem, TFilters>;
  private readonly getRowId?: (row: TItem) => DataTableRowId;
  private abortController: AbortController | null = null;

  constructor(options: DataTableControllerOptions<TItem, TFilters> = {}) {
    this.source = options.source;
    this.getRowId = options.getRowId;
    this.rows = options.initialRows ?? [];
    this.totalRows = options.initialTotalRows ?? this.rows.length;
    this.query = normalizeDataTableQuery(options.initialQuery);
    this.totalPages = Math.max(1, Math.ceil(this.totalRows / this.query.pageSize));

    if (options.autoLoad) {
      void this.load();
    }
  }

  async load(): Promise<void> {
    if (!this.source) return;

    this.abortController?.abort();
    const controller = new AbortController();
    this.abortController = controller;
    this.loading = true;
    this.error = null;

    try {
      const result = await this.source.load(this.createLoadQuery(), controller.signal);
      if (this.abortController !== controller) return;

      this.rows = result.rows;
      this.totalRows = result.totalRows;
      this.totalPages =
        result.totalPages ?? Math.max(1, Math.ceil(result.totalRows / this.query.pageSize));
      this.query = {
        ...this.query,
        page: normalizePage(result.page ?? this.query.page),
        pageSize: normalizePageSize(result.pageSize ?? this.query.pageSize)
      };
      this.pruneSelection();
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      if (this.abortController !== controller) return;

      this.error = error instanceof Error ? error.message : String(error);
    } finally {
      if (this.abortController === controller) {
        this.loading = false;
      }
    }
  }

  async reload(): Promise<void> {
    await this.load();
  }

  async setSearch(search: string): Promise<void> {
    this.updateQuery({ search, page: 1 });
    await this.load();
  }

  async setSort(sort?: DataTableSortState): Promise<void> {
    this.updateQuery({ sort, page: 1 });
    await this.load();
  }

  async setPage(page: number): Promise<void> {
    this.updateQuery({ page: normalizePage(page) });
    await this.load();
  }

  async setPageSize(pageSize: number): Promise<void> {
    this.updateQuery({ pageSize: normalizePageSize(pageSize), page: 1 });
    await this.load();
  }

  async setFilters(filters: TFilters): Promise<void> {
    this.updateQuery({ filters, page: 1 });
    await this.load();
  }

  setColumnVisibility(columnVisibility: Record<string, boolean>): void {
    this.updateQuery({ columnVisibility });
  }

  setSelection(ids: Set<DataTableRowId>): void {
    this.selectedIds = new Set(ids);
  }

  clearSelection(): void {
    this.selectedIds = new Set();
  }

  setExpanded(ids: Set<DataTableRowId>): void {
    this.expandedIds = new Set(ids);
  }

  dispose(): void {
    this.abortController?.abort();
  }

  private updateQuery(patch: Partial<DataTableQueryState<TFilters>>): void {
    this.query = {
      ...this.query,
      ...patch,
      filters: patch.filters ?? this.query.filters,
      columnVisibility: patch.columnVisibility ?? this.query.columnVisibility
    };
  }

  private createLoadQuery(): DataTableLoadQuery<TFilters> {
    return {
      search: this.query.search,
      sort: this.query.sort,
      pagination: {
        page: this.query.page,
        pageSize: this.query.pageSize
      },
      filters: this.query.filters
    };
  }

  private pruneSelection(): void {
    if (!this.getRowId || this.selectedIds.size === 0) return;

    const liveIds = new Set(this.rows.map((row) => this.getRowId?.(row) ?? ''));
    const nextSelection = new Set([...this.selectedIds].filter((id) => liveIds.has(id)));

    if (nextSelection.size !== this.selectedIds.size) {
      this.selectedIds = nextSelection;
    }
  }
}
