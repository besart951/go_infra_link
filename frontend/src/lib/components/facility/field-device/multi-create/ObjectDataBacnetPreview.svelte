<script lang="ts">
  import BacnetObjectRow from '$lib/components/facility/bacnet/BacnetObjectRow.svelte';
  import type { ObjectData } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    objectData: ObjectData | null;
    loading: boolean;
    error?: string;
  };

  let { objectData, loading, error = '' }: Props = $props();

  const t = createTranslator();
  const bacnetObjects = $derived(objectData?.bacnet_objects ?? []);
  let isExpanded = $state(false);
  let lastObjectDataId = $state<string | null>(null);

  $effect(() => {
    const currentId = objectData?.id ?? null;
    if (currentId !== lastObjectDataId) {
      isExpanded = false;
      lastObjectDataId = currentId;
    }
  });
</script>

{#if objectData}
  <div class="mt-4">
    <details class="rounded-md border bg-muted/20" bind:open={isExpanded}>
      <summary class="cursor-pointer list-none px-4 py-3 text-sm font-medium">
        <div class="flex items-center justify-between gap-3">
          <span>{$t('field_device.multi_create.object_data_preview.title')}</span>
          <span class="text-xs text-muted-foreground">
            {$t('field_device.multi_create.object_data_preview.count', {
              count: bacnetObjects.length
            })}
          </span>
        </div>
      </summary>

      <div class="space-y-3 border-t p-4">
        {#if loading}
          <p class="text-sm text-muted-foreground">
            {$t('field_device.multi_create.object_data_preview.loading')}
          </p>
        {:else if error}
          <p class="text-sm text-destructive">{error}</p>
        {:else if bacnetObjects.length === 0}
          <p class="text-sm text-muted-foreground">
            {$t('field_device.multi_create.object_data_preview.empty')}
          </p>
        {:else}
          <div class="space-y-3">
            {#each bacnetObjects as obj, index (obj.id)}
              <BacnetObjectRow {index} {obj} readOnly />
            {/each}
          </div>
        {/if}
      </div>
    </details>
  </div>
{/if}
