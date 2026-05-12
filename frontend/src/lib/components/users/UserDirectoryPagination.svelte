<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import type { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';

  type Props = {
    state: UserDirectoryPageState;
  };

  let { state }: Props = $props();
  const t = createTranslator();
</script>

{#if state.query.totalPages > 1}
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <div class="text-sm text-muted-foreground">
      {$t('messages.page_of')
        .replace('{page}', String(state.query.page))
        .replace('{total}', String(state.query.totalPages))}
      • {$t('messages.total_items').replace('{count}', String(state.query.total))}
    </div>
    <div class="flex items-center gap-2 sm:justify-end">
      <Button
        variant="outline"
        size="sm"
        disabled={!state.query.canGoToPreviousPage || state.query.loading}
        onclick={() => void state.goToPage(state.query.page - 1)}
      >
        <ChevronLeft class="mr-1 h-4 w-4" />
        {$t('messages.previous')}
      </Button>
      <Button
        variant="outline"
        size="sm"
        disabled={!state.query.canGoToNextPage || state.query.loading}
        onclick={() => void state.goToPage(state.query.page + 1)}
      >
        {$t('messages.next')}
        <ChevronRight class="ml-1 h-4 w-4" />
      </Button>
    </div>
  </div>
{/if}
