export type TableFilterRecord = Record<string, string | undefined>;
export type SortOrder = 'asc' | 'desc';

export interface DataTableQuery<TFilters extends TableFilterRecord> {
  page: number;
  pageSize: number;
  searchText: string;
  orderBy?: string;
  order?: SortOrder;
  filters: TFilters;
}

export interface DataTablePage<TItem> {
  items: TItem[];
  total: number;
  page: number;
  totalPages: number;
}

export interface DataTableFetchStrategy<TItem, TFilters extends TableFilterRecord> {
  fetch(query: DataTableQuery<TFilters>, signal?: AbortSignal): Promise<DataTablePage<TItem>>;
}

export interface BaseDataTableStateOptions<TFilters extends TableFilterRecord> {
  pageSize?: number;
  initialFilters?: TFilters;
}
