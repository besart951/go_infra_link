export type PaginatedSearchStateOptions = {
  pageSize?: number;
  initialTotalPages?: number;
  initialLoading?: boolean;
};

export type PaginatedSearchResult = {
  total: number;
  page: number;
  totalPages: number;
  searchText?: string;
};

export type PaginatedSearchSnapshot = {
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  searchText: string;
  loading: boolean;
  error: string | null;
};

export class PaginatedSearchState {
  readonly pageSize: number;

  total = $state(0);
  page = $state(1);
  totalPages = $state(0);
  searchText = $state('');
  loading = $state(false);
  error = $state<string | null>(null);

  readonly canGoToPreviousPage = $derived.by(() => this.page > 1);
  readonly canGoToNextPage = $derived.by(() => this.page < this.totalPages);
  readonly hasSearch = $derived.by(() => this.searchText.trim().length > 0);

  constructor(options: PaginatedSearchStateOptions = {}) {
    this.pageSize = options.pageSize ?? 10;
    this.totalPages = options.initialTotalPages ?? 0;
    this.loading = options.initialLoading ?? false;
  }

  snapshot(): PaginatedSearchSnapshot {
    return {
      total: this.total,
      page: this.page,
      pageSize: this.pageSize,
      totalPages: this.totalPages,
      searchText: this.searchText,
      loading: this.loading,
      error: this.error
    };
  }

  setLoading(loading: boolean): void {
    this.loading = loading;
  }

  clearError(): void {
    this.error = null;
  }

  setError(error: string | null): void {
    this.error = error;
  }

  setSearchText(searchText: string): void {
    this.searchText = searchText;
  }

  applySearch(searchText: string): void {
    this.searchText = searchText;
    this.page = 1;
  }

  goToPage(page: number): void {
    this.page = this.normalizePage(page);
  }

  goToPreviousPage(): void {
    if (!this.canGoToPreviousPage) return;
    this.goToPage(this.page - 1);
  }

  goToNextPage(): void {
    if (!this.canGoToNextPage) return;
    this.goToPage(this.page + 1);
  }

  applyResult(result: PaginatedSearchResult): void {
    this.total = Math.max(0, result.total);
    this.totalPages = Math.max(0, result.totalPages);
    this.page = this.normalizePage(result.page);
    if (result.searchText !== undefined) {
      this.searchText = result.searchText;
    }
    this.error = null;
  }

  reset(): void {
    this.total = 0;
    this.page = 1;
    this.totalPages = 0;
    this.searchText = '';
    this.loading = false;
    this.error = null;
  }

  private normalizePage(page: number): number {
    if (!Number.isFinite(page)) return 1;
    return Math.max(1, Math.floor(page));
  }
}
