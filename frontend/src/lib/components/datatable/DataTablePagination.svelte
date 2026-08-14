<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';

  interface Props {
    page: number;
    pageSize: number;
    totalRows: number;
    totalPages: number;
    pageSizeOptions?: number[];
    loading?: boolean;
    labels: {
      page: string;
      rowsPerPage: string;
      previousPage: string;
      nextPage: string;
    };
    onPageChange: (page: number) => void;
    onPageSizeChange: (pageSize: number) => void;
  }

  let {
    page,
    pageSize,
    totalRows,
    totalPages,
    pageSizeOptions = [10, 25, 50, 100],
    loading = false,
    labels,
    onPageChange,
    onPageSizeChange
  }: Props = $props();

  const firstRow = $derived(totalRows === 0 ? 0 : (page - 1) * pageSize + 1);
  const lastRow = $derived(Math.min(page * pageSize, totalRows));

  function handlePageSizeChange(event: Event): void {
    const target = event.currentTarget;
    if (!(target instanceof HTMLSelectElement)) return;
    onPageSizeChange(Number(target.value));
  }
</script>

<div class="flex flex-col gap-3 text-sm sm:flex-row sm:items-center sm:justify-between">
  <div class="text-muted-foreground">
    {labels.page
      .replace('{page}', String(page))
      .replace('{totalPages}', String(totalPages))
      .replace('{firstRow}', String(firstRow))
      .replace('{lastRow}', String(lastRow))
      .replace('{totalRows}', String(totalRows))}
  </div>

  <div class="flex flex-wrap items-center gap-2 sm:justify-end">
    <label class="flex items-center gap-2 text-muted-foreground">
      <span>{labels.rowsPerPage}</span>
      <select
        class="h-8 rounded-md border border-input bg-background px-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:opacity-50"
        value={pageSize}
        disabled={loading}
        onchange={handlePageSizeChange}
      >
        {#each pageSizeOptions as option (option)}
          <option value={option}>{option}</option>
        {/each}
      </select>
    </label>

    <Button
      variant="outline"
      size="icon-sm"
      disabled={page <= 1 || loading}
      aria-label={labels.previousPage}
      title={labels.previousPage}
      onclick={() => onPageChange(page - 1)}
    >
      <ChevronLeft class="size-4" />
    </Button>
    <Button
      variant="outline"
      size="icon-sm"
      disabled={page >= totalPages || loading}
      aria-label={labels.nextPage}
      title={labels.nextPage}
      onclick={() => onPageChange(page + 1)}
    >
      <ChevronRight class="size-4" />
    </Button>
  </div>
</div>
