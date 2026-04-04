<script lang="ts">
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { AlertCircle, Save, Undo } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useFieldDeviceState();
</script>

{#if state.editing.hasUnsavedChanges}
  <div
    class="fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 items-center gap-2 rounded-lg border bg-card p-2 shadow-lg"
  >
    <Tooltip.Root>
      <Tooltip.Trigger class="inline-flex">
        <div class="inline-flex items-center gap-1 rounded-md border px-2 py-1 text-xs font-medium">
          <AlertCircle class="h-3.5 w-3.5" />
          <span>{state.editing.pendingCount}</span>
        </div>
      </Tooltip.Trigger>
      <Tooltip.Content>
        <div class="text-sm">
          {$t('field_device.save_bar.unsaved', { count: state.editing.pendingCount })}
        </div>
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root>
      <Tooltip.Trigger class="inline-flex">
        <div
          class="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border"
          role="button"
          tabindex="0"
          onclick={() => state.savePendingEdits()}
          onkeydown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') state.savePendingEdits();
          }}
          aria-label={$t('field_device.save_bar.save_all')}
        >
          <Save class="h-4 w-4" />
        </div>
      </Tooltip.Trigger>
      <Tooltip.Content>
        <div class="text-sm">{$t('field_device.save_bar.save_all')}</div>
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root>
      <Tooltip.Trigger class="inline-flex">
        <div
          class="inline-flex h-8 w-8 cursor-pointer items-center justify-center rounded-md"
          role="button"
          tabindex="0"
          onclick={() => state.discardPendingEdits()}
          onkeydown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') state.discardPendingEdits();
          }}
          aria-label={$t('field_device.save_bar.discard')}
        >
          <Undo class="h-4 w-4" />
        </div>
      </Tooltip.Trigger>
      <Tooltip.Content>
        <div class="text-sm">{$t('field_device.save_bar.discard')}</div>
      </Tooltip.Content>
    </Tooltip.Root>
  </div>
{/if}
