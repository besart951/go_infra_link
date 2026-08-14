<script
  lang="ts"
  generics="TItem, TFilters extends import('./types.js').DataTableFilterRecord = import('./types.js').DataTableFilterRecord"
>
  import type { Snippet } from 'svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { keyboardTableNavigation } from '$lib/actions/keyboardTableNavigation.js';
  import { cn } from '$lib/utils.js';
  import ArrowDown from '@lucide/svelte/icons/arrow-down';
  import ArrowUp from '@lucide/svelte/icons/arrow-up';
  import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Search from '@lucide/svelte/icons/search';
  import DataTableColumnVisibilityMenu from './DataTableColumnVisibilityMenu.svelte';
  import DataTablePagination from './DataTablePagination.svelte';
  import {
    formatCellValue,
    getColumnValue,
    getVisibleColumns,
    nextSortState,
    normalizeDataTableQuery,
    processDataTableRows
  } from './dataTableModel.js';
  import type {
    DataTableCellContext,
    DataTableColumn,
    DataTableColumnVisibilityState,
    DataTableDensity,
    DataTableEmptyState,
    DataTableLabels,
    DataTableManualMode,
    DataTablePaginationOptions,
    DataTableQueryState,
    DataTableRowContext,
    DataTableRowId,
    DataTableSelectionOptions,
    DataTableToolbarContext
  } from './types.js';

  interface Props {
    rows: TItem[];
    columns: DataTableColumn<TItem>[];
    getRowId?: (row: TItem) => DataTableRowId;
    query?: Partial<DataTableQueryState<TFilters>>;
    selectedIds?: Set<DataTableRowId>;
    expandedIds?: Set<DataTableRowId>;
    manual?: DataTableManualMode;
    pagination?: boolean | DataTablePaginationOptions;
    selectable?: boolean | DataTableSelectionOptions<TItem>;
    loading?: boolean;
    error?: string | null;
    density?: DataTableDensity;
    labels?: Partial<DataTableLabels>;
    empty?: DataTableEmptyState;
    searchPlaceholder?: string;
    enableSearch?: boolean;
    enableColumnVisibility?: boolean;
    tableClass?: string;
    wrapperClass?: string;
    rowClass?: string | ((context: DataTableRowContext<TItem>) => string | undefined);
    rowCells?: Snippet<[TItem]>;
    toolbar?: Snippet<[DataTableToolbarContext<TItem, TFilters>]>;
    filters?: Snippet<[DataTableToolbarContext<TItem, TFilters>]>;
    rowActions?: Snippet<[DataTableRowContext<TItem>]>;
    expandedRow?: Snippet<[DataTableRowContext<TItem>]>;
    onQueryChange?: (query: DataTableQueryState<TFilters>) => void;
    onSelectionChange?: (ids: Set<DataTableRowId>, rows: TItem[]) => void;
    onExpandedChange?: (ids: Set<DataTableRowId>) => void;
    onRowClick?: (context: DataTableRowContext<TItem>) => void;
    reload?: () => void | Promise<void>;
  }

  let {
    rows,
    columns,
    getRowId = defaultGetRowId,
    query,
    selectedIds,
    expandedIds,
    manual = {},
    pagination = true,
    selectable = false,
    loading = false,
    error = null,
    density = 'comfortable',
    labels,
    empty,
    searchPlaceholder,
    enableSearch = true,
    enableColumnVisibility = true,
    tableClass,
    wrapperClass,
    rowClass,
    rowCells,
    toolbar,
    filters,
    rowActions,
    expandedRow,
    onQueryChange,
    onSelectionChange,
    onExpandedChange,
    onRowClick,
    reload
  }: Props = $props();

  let localQuery = $state<DataTableQueryState<TFilters>>(
    normalizeDataTableQuery<TFilters>(undefined)
  );
  const localSelectedIds = new SvelteSet<DataTableRowId>();
  const localExpandedIds = new SvelteSet<DataTableRowId>();

  const resolvedLabels = $derived<DataTableLabels>({
    search: 'Search',
    columns: 'Columns',
    noResults: 'No rows found',
    noResultsDescription: 'Adjust the search or filters and try again.',
    loading: 'Loading rows',
    error: 'Unable to load rows',
    page: '{firstRow}-{lastRow} of {totalRows}',
    rowsPerPage: 'Rows',
    selectedRows: '{count} selected',
    previousPage: 'Previous page',
    nextPage: 'Next page',
    sortAscending: 'Sort ascending',
    sortDescending: 'Sort descending',
    clearSort: 'Clear sorting',
    expandRow: 'Expand row',
    collapseRow: 'Collapse row',
    ...labels
  });

  const paginationOptions = $derived(
    typeof pagination === 'object'
      ? pagination
      : {
          enabled: pagination
        }
  );
  const paginationEnabled = $derived(paginationOptions.enabled !== false);
  const selectionOptions = $derived(
    typeof selectable === 'object'
      ? selectable
      : {
          enabled: selectable
        }
  );
  const selectionEnabled = $derived(selectionOptions.enabled === true);
  const queryState = $derived(normalizeDataTableQuery<TFilters>(query ?? localQuery));
  const currentSelectedIds = $derived(selectedIds ?? localSelectedIds);
  const currentExpandedIds = $derived(expandedIds ?? localExpandedIds);
  const visibleColumns = $derived(getVisibleColumns(columns, queryState.columnVisibility));
  const processedRows = $derived(
    processDataTableRows(
      rows,
      columns,
      queryState,
      {
        ...manual,
        pagination: manual.pagination || !paginationEnabled
      },
      paginationOptions.totalRows,
      paginationOptions.totalPages
    )
  );
  const visibleRows = $derived(processedRows.rows);
  const visibleRowContexts = $derived(
    visibleRows.map((row, rowIndex) => createRowContext(row, rowIndex))
  );
  const selectableRowContexts = $derived(
    visibleRowContexts.filter((context) => !isSelectionDisabled(context.row))
  );
  const allVisibleRowsSelected = $derived(
    selectableRowContexts.length > 0 &&
      selectableRowContexts.every((context) => currentSelectedIds.has(context.rowId))
  );
  const someVisibleRowsSelected = $derived(
    selectableRowContexts.some((context) => currentSelectedIds.has(context.rowId)) &&
      !allVisibleRowsSelected
  );
  const columnCount = $derived(
    visibleColumns.length +
      (selectionEnabled ? 1 : 0) +
      (expandedRow ? 1 : 0) +
      (rowActions ? 1 : 0)
  );
  const selectedRows = $derived(rows.filter((row) => currentSelectedIds.has(getRowId(row))));
  const toolbarContext = $derived<DataTableToolbarContext<TItem, TFilters>>({
    query: queryState,
    rows,
    visibleRows,
    selectedIds: currentSelectedIds,
    selectedRows,
    clearSelection,
    reload
  });
  const densityClass = $derived(getDensityClass(density));
  const resolvedTableClass = $derived(cn(densityClass, tableClass));
  const emptyTitle = $derived(empty?.title ?? resolvedLabels.noResults);
  const emptyDescription = $derived(empty?.description ?? resolvedLabels.noResultsDescription);
  const hasToolbar = $derived(enableSearch || toolbar || filters || enableColumnVisibility);

  function defaultGetRowId(row: TItem): DataTableRowId {
    const candidate = (row as { id?: unknown }).id;
    if (typeof candidate === 'string' || typeof candidate === 'number') return String(candidate);
    return String(rows.indexOf(row));
  }

  function emitQueryChange(nextQuery: DataTableQueryState<TFilters>): void {
    localQuery = nextQuery;
    onQueryChange?.(nextQuery);
  }

  function patchQuery(patch: Partial<DataTableQueryState<TFilters>>): void {
    emitQueryChange({
      ...queryState,
      ...patch,
      filters: patch.filters ?? queryState.filters,
      columnVisibility: patch.columnVisibility ?? queryState.columnVisibility
    });
  }

  function handleSearchInput(event: Event): void {
    const target = event.currentTarget;
    if (!(target instanceof HTMLInputElement)) return;
    patchQuery({ search: target.value, page: 1 });
  }

  function handleSort(column: DataTableColumn<TItem>): void {
    if (!column.sortable) return;
    patchQuery({
      sort: nextSortState(queryState.sort, column.key),
      page: 1
    });
  }

  function handleColumnVisibilityChange(columnVisibility: DataTableColumnVisibilityState): void {
    patchQuery({ columnVisibility });
  }

  function handlePageChange(page: number): void {
    patchQuery({ page });
  }

  function handlePageSizeChange(pageSize: number): void {
    patchQuery({ pageSize, page: 1 });
  }

  function emitSelectionChange(nextIds: Set<DataTableRowId>): void {
    localSelectedIds.clear();
    for (const id of nextIds) {
      localSelectedIds.add(id);
    }

    onSelectionChange?.(
      nextIds,
      rows.filter((row) => nextIds.has(getRowId(row)))
    );
  }

  function toggleRowSelection(context: DataTableRowContext<TItem>): void {
    if (isSelectionDisabled(context.row)) return;

    const nextIds = new SvelteSet(currentSelectedIds);
    if (nextIds.has(context.rowId)) {
      nextIds.delete(context.rowId);
    } else {
      nextIds.add(context.rowId);
    }
    emitSelectionChange(nextIds);
  }

  function toggleVisibleRowsSelection(): void {
    const nextIds = new SvelteSet(currentSelectedIds);

    if (allVisibleRowsSelected) {
      for (const context of selectableRowContexts) {
        nextIds.delete(context.rowId);
      }
    } else {
      for (const context of selectableRowContexts) {
        nextIds.add(context.rowId);
      }
    }

    emitSelectionChange(nextIds);
  }

  function clearSelection(): void {
    emitSelectionChange(new SvelteSet());
  }

  function emitExpandedChange(nextIds: Set<DataTableRowId>): void {
    localExpandedIds.clear();
    for (const id of nextIds) {
      localExpandedIds.add(id);
    }

    onExpandedChange?.(nextIds);
  }

  function toggleExpandedRow(context: DataTableRowContext<TItem>): void {
    const nextIds = new SvelteSet(currentExpandedIds);
    if (nextIds.has(context.rowId)) {
      nextIds.delete(context.rowId);
    } else {
      nextIds.add(context.rowId);
    }
    emitExpandedChange(nextIds);
  }

  function createRowContext(row: TItem, rowIndex: number): DataTableRowContext<TItem> {
    const rowId = getRowId(row);
    return {
      row,
      rowId,
      rowIndex,
      selected: currentSelectedIds.has(rowId),
      expanded: currentExpandedIds.has(rowId)
    };
  }

  function createCellContext(
    context: DataTableRowContext<TItem>,
    column: DataTableColumn<TItem>
  ): DataTableCellContext<TItem> {
    return {
      row: context.row,
      rowId: context.rowId,
      rowIndex: context.rowIndex,
      column,
      value: getColumnValue(context.row, column)
    };
  }

  function isSelectionDisabled(row: TItem): boolean {
    return selectionOptions.getDisabled?.(row) ?? false;
  }

  function getRowClass(context: DataTableRowContext<TItem>): string | undefined {
    if (typeof rowClass === 'function') return rowClass(context);
    return rowClass;
  }

  function getSortLabel(column: DataTableColumn<TItem>): string {
    if (!queryState.sort || queryState.sort.key !== column.key) return resolvedLabels.sortAscending;
    return queryState.sort.direction === 'asc'
      ? resolvedLabels.sortDescending
      : resolvedLabels.clearSort;
  }

  function getHeaderClass(column: DataTableColumn<TItem>): string {
    return cn(
      column.align === 'right' && 'text-right',
      column.align === 'center' && 'text-center',
      column.class,
      column.headerClass
    );
  }

  function getCellClass(column: DataTableColumn<TItem>): string {
    return cn(
      'max-w-80 overflow-hidden',
      column.align === 'right' && 'text-right',
      column.align === 'center' && 'text-center',
      column.class,
      column.cellClass
    );
  }

  function getDensityClass(value: DataTableDensity): string {
    switch (value) {
      case 'compact':
        return 'text-xs [&_td]:px-2 [&_td]:py-1 [&_th]:h-8 [&_th]:px-2 [&_th]:py-1';
      case 'spacious':
        return 'text-sm [&_td]:px-4 [&_td]:py-3 [&_th]:h-12 [&_th]:px-4 [&_th]:py-3';
      case 'comfortable':
      default:
        return 'text-sm [&_td]:px-3 [&_td]:py-2 [&_th]:h-10 [&_th]:px-3 [&_th]:py-2';
    }
  }
</script>

<div class={cn('flex min-w-0 flex-col gap-3', wrapperClass)}>
  {#if hasToolbar}
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex min-w-0 flex-1 flex-col gap-2 sm:flex-row sm:items-center">
        {#if enableSearch}
          <div class="relative min-w-0 flex-1">
            <Search
              class="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              type="search"
              value={queryState.search}
              placeholder={searchPlaceholder ?? resolvedLabels.search}
              aria-label={searchPlaceholder ?? resolvedLabels.search}
              class="pl-9"
              oninput={handleSearchInput}
            />
          </div>
        {/if}

        {@render filters?.(toolbarContext)}
      </div>

      <div class="flex flex-wrap items-center gap-2 sm:justify-end">
        {@render toolbar?.(toolbarContext)}

        {#if currentSelectedIds.size > 0}
          <span class="text-sm whitespace-nowrap text-muted-foreground">
            {resolvedLabels.selectedRows.replace('{count}', String(currentSelectedIds.size))}
          </span>
        {/if}

        {#if enableColumnVisibility}
          <DataTableColumnVisibilityMenu
            columns={columns as DataTableColumn<unknown>[]}
            columnVisibility={queryState.columnVisibility}
            label={resolvedLabels.columns}
            onChange={handleColumnVisibilityChange}
          />
        {/if}
      </div>
    </div>
  {/if}

  {#if error}
    <div
      role="alert"
      class="rounded-md border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive"
    >
      <p class="font-medium">{resolvedLabels.error}</p>
      <p>{error}</p>
    </div>
  {/if}

  <div
    use:keyboardTableNavigation
    class="max-w-full min-w-0 overflow-hidden rounded-xl border bg-card shadow-sm"
  >
    <Table.Root class={cn('w-max min-w-full table-auto', resolvedTableClass)}>
      <Table.Header>
        <Table.Row>
          {#if selectionEnabled}
            <Table.Head resizable={false} class="w-9 max-w-9 min-w-9 px-0!">
              <div class="flex justify-center">
                <Checkbox
                  checked={allVisibleRowsSelected}
                  indeterminate={someVisibleRowsSelected}
                  disabled={selectableRowContexts.length === 0}
                  aria-label={selectionOptions.selectAllAriaLabel ?? 'Select all visible rows'}
                  onCheckedChange={toggleVisibleRowsSelection}
                />
              </div>
            </Table.Head>
          {/if}

          {#if expandedRow}
            <Table.Head resizable={false} class="w-9 max-w-9 min-w-9 px-0!"></Table.Head>
          {/if}

          {#each visibleColumns as column (column.key)}
            <Table.Head
              class={getHeaderClass(column)}
              resizable={column.resizable}
              minResizeWidth={column.minResizeWidth}
              aria-sort={queryState.sort?.key === column.key
                ? queryState.sort.direction === 'asc'
                  ? 'ascending'
                  : 'descending'
                : undefined}
            >
              {#if column.header}
                {@render column.header(column)}
              {:else if column.sortable}
                <Button
                  type="button"
                  variant="ghost"
                  class="h-auto w-full min-w-0 justify-start p-0 text-left font-medium underline-offset-4 hover:underline"
                  aria-label={`${getSortLabel(column)}: ${column.label}`}
                  onclick={() => handleSort(column)}
                >
                  <span class="min-w-0 truncate">{column.label}</span>
                  {#if queryState.sort?.key === column.key && queryState.sort.direction === 'asc'}
                    <ArrowUp class="size-3.5" />
                  {:else if queryState.sort?.key === column.key && queryState.sort.direction === 'desc'}
                    <ArrowDown class="size-3.5" />
                  {:else}
                    <ChevronsUpDown class="size-3.5 text-muted-foreground" />
                  {/if}
                </Button>
              {:else}
                {column.label}
              {/if}
            </Table.Head>
          {/each}

          {#if rowActions}
            <Table.Head resizable={false} class="w-16 max-w-16 min-w-16"></Table.Head>
          {/if}
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {#if loading && visibleRows.length === 0}
          <Table.LoadingRows loading {columnCount} rowCount={8} delayMs={0} />
        {:else if visibleRows.length === 0}
          <Table.Row>
            <Table.Cell colspan={columnCount} class="h-28 text-center">
              <div class="flex flex-col items-center justify-center gap-1 text-muted-foreground">
                <p class="font-medium">{emptyTitle}</p>
                {#if queryState.search || empty?.description}
                  <p class="text-sm">{emptyDescription}</p>
                {/if}
              </div>
            </Table.Cell>
          </Table.Row>
        {:else}
          {#each visibleRowContexts as context (context.rowId)}
            <Table.Row
              data-state={context.selected ? 'selected' : undefined}
              class={cn(
                loading && 'opacity-60',
                onRowClick && 'cursor-pointer',
                getRowClass(context)
              )}
              onclick={() => onRowClick?.(context)}
            >
              {#if selectionEnabled}
                <Table.Cell class="w-9 max-w-9 min-w-9 px-0!">
                  <div class="flex justify-center">
                    <Checkbox
                      checked={context.selected}
                      disabled={isSelectionDisabled(context.row)}
                      aria-label={selectionOptions.ariaLabel ?? 'Select row'}
                      onCheckedChange={() => toggleRowSelection(context)}
                    />
                  </div>
                </Table.Cell>
              {/if}

              {#if expandedRow}
                <Table.Cell class="w-9 max-w-9 min-w-9 px-0!">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-expanded={context.expanded}
                    aria-label={context.expanded
                      ? resolvedLabels.collapseRow
                      : resolvedLabels.expandRow}
                    title={context.expanded ? resolvedLabels.collapseRow : resolvedLabels.expandRow}
                    onclick={(event) => {
                      event.stopPropagation();
                      toggleExpandedRow(context);
                    }}
                  >
                    {#if context.expanded}
                      <ChevronDown class="size-4" />
                    {:else}
                      <ChevronRight class="size-4" />
                    {/if}
                  </Button>
                </Table.Cell>
              {/if}

              {#if rowCells}
                {@render rowCells(context.row)}
              {:else}
                {#each visibleColumns as column (column.key)}
                  <Table.Cell class={getCellClass(column)}>
                    {#if column.cell}
                      {@render column.cell(createCellContext(context, column))}
                    {:else}
                      <span class="block truncate">{formatCellValue(context.row, column)}</span>
                    {/if}
                  </Table.Cell>
                {/each}
              {/if}

              {#if rowActions}
                <Table.Cell class="w-16 max-w-16 min-w-16 text-right">
                  {@render rowActions(context)}
                </Table.Cell>
              {/if}
            </Table.Row>

            {#if expandedRow && context.expanded}
              <Table.Row class="bg-muted/30 hover:bg-muted/40">
                <Table.Cell colspan={columnCount} class="p-0">
                  {@render expandedRow(context)}
                </Table.Cell>
              </Table.Row>
            {/if}
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>

  {#if paginationEnabled && processedRows.totalPages > 1}
    <DataTablePagination
      page={queryState.page}
      pageSize={queryState.pageSize}
      totalRows={processedRows.totalRows}
      totalPages={processedRows.totalPages}
      pageSizeOptions={paginationOptions.pageSizeOptions}
      {loading}
      labels={{
        page: resolvedLabels.page,
        rowsPerPage: resolvedLabels.rowsPerPage,
        previousPage: resolvedLabels.previousPage,
        nextPage: resolvedLabels.nextPage
      }}
      onPageChange={handlePageChange}
      onPageSizeChange={handlePageSizeChange}
    />
  {/if}
</div>
