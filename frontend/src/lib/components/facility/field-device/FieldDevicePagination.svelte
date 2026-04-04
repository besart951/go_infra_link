<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { ChevronLeft, ChevronRight } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useFieldDeviceState();
</script>

{#if state.totalPages > 1}
  <div class="flex items-center justify-between">
    <div class="text-sm text-muted-foreground">
      {$t('messages.page_of', { page: state.page, total: state.totalPages })}
      &bull; {$t('messages.total_items', { count: state.total })}
    </div>
    <div class="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={state.page <= 1 || state.loading}
        onclick={() => void state.goToPreviousPage()}
      >
        <ChevronLeft class="mr-1 h-4 w-4" />
        {$t('common.previous')}
      </Button>
      <Button
        variant="outline"
        size="sm"
        disabled={state.page >= state.totalPages || state.loading}
        onclick={() => void state.goToNextPage()}
      >
        {$t('common.next')}
        <ChevronRight class="ml-1 h-4 w-4" />
      </Button>
    </div>
  </div>
{/if}
