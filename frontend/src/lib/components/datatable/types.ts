import type { Snippet } from 'svelte';

export type DataTableRowId = string;
export type DataTableSortDirection = 'asc' | 'desc';
export type DataTableDensity = 'compact' | 'comfortable' | 'spacious';
export type DataTableFilterRecord = Record<string, string | number | boolean | null | undefined>;
export type DataTableColumnVisibilityState = Record<string, boolean>;

export interface DataTableSortState {
  key: string;
  direction: DataTableSortDirection;
}

export interface DataTableQueryState<
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  search: string;
  sort?: DataTableSortState;
  page: number;
  pageSize: number;
  filters: TFilters;
  columnVisibility: DataTableColumnVisibilityState;
}

export interface DataTableLoadQuery<
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  search: string;
  sort?: DataTableSortState;
  pagination: {
    page: number;
    pageSize: number;
  };
  filters: TFilters;
}

export interface DataTableLoadResult<TItem> {
  rows: TItem[];
  totalRows: number;
  page?: number;
  pageSize?: number;
  totalPages?: number;
}

export interface DataTableSource<
  TItem,
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  load(
    query: DataTableLoadQuery<TFilters>,
    signal?: AbortSignal
  ): Promise<DataTableLoadResult<TItem>>;
}

export interface DataTableColumn<TItem> {
  key: string;
  label: string;
  accessor?: keyof TItem | ((row: TItem) => unknown);
  cell?: Snippet<[DataTableCellContext<TItem>]>;
  header?: Snippet<[DataTableColumn<TItem>]>;
  sortable?: boolean;
  searchable?: boolean;
  hideable?: boolean;
  defaultVisible?: boolean;
  resizable?: boolean;
  minResizeWidth?: number;
  align?: 'left' | 'center' | 'right';
  headerClass?: string;
  cellClass?: string;
  class?: string;
  compare?: (left: TItem, right: TItem) => number;
  searchValue?: (row: TItem) => string;
  format?: (value: unknown, row: TItem) => string;
}

export interface DataTableCellContext<TItem> {
  row: TItem;
  rowId: DataTableRowId;
  column: DataTableColumn<TItem>;
  value: unknown;
  rowIndex: number;
}

export interface DataTableRowContext<TItem> {
  row: TItem;
  rowId: DataTableRowId;
  rowIndex: number;
  selected: boolean;
  expanded: boolean;
}

export interface DataTableToolbarContext<
  TItem,
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
> {
  query: DataTableQueryState<TFilters>;
  rows: TItem[];
  visibleRows: TItem[];
  selectedIds: Set<DataTableRowId>;
  selectedRows: TItem[];
  clearSelection(): void;
  reload?: () => void | Promise<void>;
}

export interface DataTableManualMode {
  search?: boolean;
  sort?: boolean;
  pagination?: boolean;
}

export interface DataTableSelectionOptions<TItem> {
  enabled?: boolean;
  getDisabled?: (row: TItem) => boolean;
  ariaLabel?: string;
  selectAllAriaLabel?: string;
}

export interface DataTablePaginationOptions {
  enabled?: boolean;
  totalRows?: number;
  totalPages?: number;
  pageSizeOptions?: number[];
}

export interface DataTableLabels {
  search: string;
  columns: string;
  noResults: string;
  noResultsDescription: string;
  loading: string;
  error: string;
  page: string;
  rowsPerPage: string;
  selectedRows: string;
  previousPage: string;
  nextPage: string;
  sortAscending: string;
  sortDescending: string;
  clearSort: string;
  expandRow: string;
  collapseRow: string;
}

export interface DataTableColumnDefinition {
  key: string;
  label: string;
  width?: string;
}

export interface DataTableEmptyState {
  title?: string;
  description?: string;
}

export interface DataTableProcessResult<TItem> {
  rows: TItem[];
  filteredRows: TItem[];
  sortedRows: TItem[];
  totalRows: number;
  totalPages: number;
}
