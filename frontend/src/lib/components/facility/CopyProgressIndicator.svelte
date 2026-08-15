<script lang="ts">
  import { createTranslator } from '$lib/i18n/translator.js';
  import { copyOperation } from '$lib/state/copyOperation.svelte.js';
  import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';

  const t = createTranslator();

  const stageKey: Record<string, string> = {
    queued: 'facility.copy_progress.queued',
    preparing: 'facility.copy_progress.preparing',
    copying_root: 'facility.copy_progress.copying_root',
    copying_controllers: 'facility.copy_progress.copying_controllers',
    copying_system_types: 'facility.copy_progress.copying_system_types',
    copying_field_devices: 'facility.copy_progress.copying_field_devices',
    finalizing: 'facility.copy_progress.finalizing',
    completed: 'facility.copy_progress.completed',
    failed: 'facility.copy_progress.failed'
  };

  const stageLabel = $derived(
    $t(stageKey[copyOperation.stage] ?? 'facility.copy_progress.preparing')
  );
</script>

{#if copyOperation.isPending}
  <div
    class="fixed right-4 bottom-4 z-50 w-[min(24rem,calc(100vw-2rem))] rounded-lg border bg-card p-4 shadow-lg"
  >
    <div class="flex items-start gap-3" aria-live="polite" aria-atomic="true">
      <LoaderCircleIcon class="mt-0.5 size-5 shrink-0 animate-spin text-primary" />
      <div class="min-w-0 flex-1">
        <div class="flex items-center justify-between gap-3">
          <p class="text-sm font-medium">{stageLabel}</p>
          <span class="text-sm text-muted-foreground tabular-nums">{copyOperation.progress}%</span>
        </div>
        <div
          class="mt-2 h-2 overflow-hidden rounded-full bg-muted"
          role="progressbar"
          aria-label={$t('facility.copy_progress.aria')}
          aria-valuemin="0"
          aria-valuemax="100"
          aria-valuenow={copyOperation.progress}
        >
          <div
            class="h-full rounded-full bg-primary transition-[width] duration-300"
            style={`width: ${copyOperation.progress}%`}
          ></div>
        </div>
        {#if copyOperation.connectionInterrupted}
          <p class="mt-2 text-xs text-muted-foreground">
            {$t('facility.copy_progress.connection_interrupted')}
          </p>
        {/if}
      </div>
    </div>
  </div>
{/if}
