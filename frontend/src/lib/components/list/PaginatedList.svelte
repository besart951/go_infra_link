<script lang="ts" generics="T">
  import { Input } from '$lib/components/ui/input/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Search, ChevronLeft, ChevronRight } from '@lucide/svelte';
  import type { Snippet } from 'svelte';
  import type { ListState } from '$lib/application/useCases/listUseCase.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  interface Props {
    state: ListState<T>;
    columns: Array<{ key: string; label: string; width?: string }>;
    rowSnippet: Snippet<[T]>;
    emptyMessage?: string;
    searchPlaceholder?: string;
    onSearch: (text: string) => void;
    onPageChange: (page: number) => void;
    onReload?: () => void;
  }

  let {
    state,
    columns,
    rowSnippet,
    emptyMessage = 'No items found',
    searchPlaceholder = 'Search...',
    onSearch,
    onPageChange,
    onReload
  }: Props = $props();

  let searchInput = $derived(state.searchText);

  function handleSearchInput(e: Event) {
    const value = (e.target as HTMLInputElement).value;
    onSearch(value);
  }

  function handlePrevious() {
    if (state.page > 1) {
      onPageChange(state.page - 1);
    }
  }

  function handleNext() {
    if (state.page < state.totalPages) {
      onPageChange(state.page + 1);
    }
  }
</script>

<div class="space-y-4">
  <!-- Search Bar -->
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
    <div class="relative min-w-0 flex-1">
      <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        placeholder={searchPlaceholder}
        aria-label={searchPlaceholder}
        class="pl-9"
        value={searchInput}
        oninput={handleSearchInput}
      />
    </div>
    {#if onReload}
      <Button class="w-full sm:w-auto" variant="outline" onclick={onReload} disabled={state.loading}
        >{$t('messages.refresh')}</Button
      >
    {/if}
  </div>

  <!-- Error Message -->
  {#if state.error}
    <div
      role="alert"
      class="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
    >
      <p class="font-medium">{$t('messages.error')}</p>
      <p>{state.error}</p>
    </div>
  {/if}

  <!-- Table -->
  <div class="overflow-hidden rounded-xl border bg-card shadow-sm">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          {#each columns as column}
            <Table.Head class={column.width}>{column.label}</Table.Head>
          {/each}
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#if state.loading && state.items.length === 0}
          <Table.LoadingRows loading columnCount={columns.length} />
        {:else if state.items.length === 0}
          <Table.Row>
            <Table.Cell colspan={columns.length} class="h-32 text-center">
              <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                <p class="font-medium">{emptyMessage}</p>
                {#if state.searchText}
                  <p class="text-sm">{$t('messages.try_adjusting_search')}</p>
                {/if}
              </div>
            </Table.Cell>
          </Table.Row>
        {:else}
          {#each state.items as item}
            <Table.Row class={state.loading ? 'opacity-60' : undefined}>
              {@render rowSnippet(item)}
            </Table.Row>
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>

  <!-- Pagination -->
  {#if state.totalPages > 1}
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="text-sm text-muted-foreground">
        {$t('messages.page_of')
          .replace('{page}', String(state.page))
          .replace('{total}', String(state.totalPages))} • {$t('messages.total_items').replace(
          '{count}',
          String(state.total)
        )}
      </div>
      <div class="flex items-center gap-2 sm:justify-end">
        <Button
          variant="outline"
          size="sm"
          disabled={state.page <= 1 || state.loading}
          onclick={handlePrevious}
        >
          <ChevronLeft class="mr-1 h-4 w-4" />
          {$t('messages.previous')}
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={state.page >= state.totalPages || state.loading}
          onclick={handleNext}
        >
          {$t('messages.next')}
          <ChevronRight class="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  {/if}
</div>
