import type {
  DataTableColumn,
  DataTableColumnVisibilityState,
  DataTableFilterRecord,
  DataTableProcessResult,
  DataTableQueryState,
  DataTableSortState
} from './types.js';

export const DEFAULT_DATA_TABLE_PAGE_SIZE = 25;

export function createDefaultDataTableQuery<
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
>(overrides: Partial<DataTableQueryState<TFilters>> = {}): DataTableQueryState<TFilters> {
  return {
    search: overrides.search ?? '',
    sort: overrides.sort,
    page: normalizePage(overrides.page),
    pageSize: normalizePageSize(overrides.pageSize),
    filters: overrides.filters ?? ({} as TFilters),
    columnVisibility: overrides.columnVisibility ?? {}
  };
}

export function normalizeDataTableQuery<
  TFilters extends DataTableFilterRecord = DataTableFilterRecord
>(query: Partial<DataTableQueryState<TFilters>> | undefined): DataTableQueryState<TFilters> {
  return createDefaultDataTableQuery(query);
}

export function getColumnValue<TItem>(row: TItem, column: DataTableColumn<TItem>): unknown {
  if (typeof column.accessor === 'function') return column.accessor(row);
  if (typeof column.accessor === 'string') return row[column.accessor];
  return (row as Record<string, unknown>)[column.key];
}

export function getVisibleColumns<TItem>(
  columns: DataTableColumn<TItem>[],
  columnVisibility: DataTableColumnVisibilityState
): DataTableColumn<TItem>[] {
  return columns.filter((column) => isColumnVisible(column, columnVisibility));
}

export function isColumnVisible<TItem>(
  column: DataTableColumn<TItem>,
  columnVisibility: DataTableColumnVisibilityState
): boolean {
  if (column.defaultVisible === false && columnVisibility[column.key] !== true) return false;
  return columnVisibility[column.key] !== false;
}

export function toggleColumnVisibility<TItem>(
  column: DataTableColumn<TItem>,
  columnVisibility: DataTableColumnVisibilityState
): DataTableColumnVisibilityState {
  const currentlyVisible = isColumnVisible(column, columnVisibility);
  return {
    ...columnVisibility,
    [column.key]: !currentlyVisible
  };
}

export function nextSortState(
  current: DataTableSortState | undefined,
  columnKey: string
): DataTableSortState | undefined {
  if (!current || current.key !== columnKey) {
    return { key: columnKey, direction: 'asc' };
  }

  if (current.direction === 'asc') {
    return { key: columnKey, direction: 'desc' };
  }

  return undefined;
}

export function processDataTableRows<TItem>(
  rows: TItem[],
  columns: DataTableColumn<TItem>[],
  query: DataTableQueryState,
  manual: { search?: boolean; sort?: boolean; pagination?: boolean } = {},
  totalRowsOverride?: number,
  totalPagesOverride?: number
): DataTableProcessResult<TItem> {
  const filteredRows = manual.search ? rows : filterRows(rows, columns, query.search);
  const sortedRows = manual.sort ? filteredRows : sortRows(filteredRows, columns, query.sort);
  const totalRows = totalRowsOverride ?? sortedRows.length;
  const totalPages =
    totalPagesOverride ?? Math.max(1, Math.ceil(totalRows / normalizePageSize(query.pageSize)));
  const pageRows = manual.pagination
    ? sortedRows
    : paginateRows(sortedRows, query.page, query.pageSize);

  return {
    rows: pageRows,
    filteredRows,
    sortedRows,
    totalRows,
    totalPages
  };
}

export function formatCellValue<TItem>(row: TItem, column: DataTableColumn<TItem>): string {
  const value = getColumnValue(row, column);
  if (column.format) return column.format(value, row);
  if (value === null || value === undefined) return '';
  if (value instanceof Date) return value.toLocaleString();
  return String(value);
}

export function normalizePage(page: number | undefined): number {
  if (!Number.isFinite(page) || !page || page < 1) return 1;
  return Math.floor(page);
}

export function normalizePageSize(pageSize: number | undefined): number {
  if (!Number.isFinite(pageSize) || !pageSize || pageSize < 1) return DEFAULT_DATA_TABLE_PAGE_SIZE;
  return Math.floor(pageSize);
}

function filterRows<TItem>(
  rows: TItem[],
  columns: DataTableColumn<TItem>[],
  search: string
): TItem[] {
  const normalizedSearch = search.trim().toLocaleLowerCase();
  if (!normalizedSearch) return rows;

  const searchableColumns = columns.filter((column) => column.searchable !== false);
  return rows.filter((row) =>
    searchableColumns.some((column) =>
      getSearchText(row, column).toLocaleLowerCase().includes(normalizedSearch)
    )
  );
}

function getSearchText<TItem>(row: TItem, column: DataTableColumn<TItem>): string {
  if (column.searchValue) return column.searchValue(row);
  const value = getColumnValue(row, column);
  if (value === null || value === undefined) return '';
  if (value instanceof Date) return value.toISOString();
  return String(value);
}

function sortRows<TItem>(
  rows: TItem[],
  columns: DataTableColumn<TItem>[],
  sort: DataTableSortState | undefined
): TItem[] {
  if (!sort) return rows;

  const column = columns.find((candidate) => candidate.key === sort.key);
  if (!column || column.sortable === false) return rows;

  const direction = sort.direction === 'desc' ? -1 : 1;
  return [...rows].sort((left, right) => {
    const result = column.compare
      ? column.compare(left, right)
      : compareValues(getColumnValue(left, column), getColumnValue(right, column));

    return result * direction;
  });
}

function compareValues(left: unknown, right: unknown): number {
  if (left === right) return 0;
  if (left === null || left === undefined || left === '') return 1;
  if (right === null || right === undefined || right === '') return -1;

  if (typeof left === 'number' && typeof right === 'number') {
    return left - right;
  }

  if (left instanceof Date && right instanceof Date) {
    return left.getTime() - right.getTime();
  }

  return String(left).localeCompare(String(right), undefined, {
    numeric: true,
    sensitivity: 'base'
  });
}

function paginateRows<TItem>(rows: TItem[], page: number, pageSize: number): TItem[] {
  const normalizedPage = normalizePage(page);
  const normalizedPageSize = normalizePageSize(pageSize);
  const start = (normalizedPage - 1) * normalizedPageSize;
  return rows.slice(start, start + normalizedPageSize);
}
