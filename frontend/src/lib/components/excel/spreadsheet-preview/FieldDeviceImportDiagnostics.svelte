<script lang="ts">
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';

  interface Props {
    view: FieldDeviceImportViewState;
  }

  let { view }: Props = $props();
  const t = createTranslator();
</script>

{#if view.diagnostics.length > 0}
  <div class="mt-4 space-y-2">
    <h4 class="text-sm font-medium">{$t('field_device.importer.diagnostics.title')}</h4>
    <div class="max-h-52 space-y-2 overflow-auto pr-1">
      {#each view.diagnostics as diagnostic (diagnostic.id)}
        <div class={`rounded-md border p-2 text-xs ${view.diagnosticClass(diagnostic)}`}>
          <div class="font-medium">
            {diagnostic.cell ? `${diagnostic.cell.address}: ` : ''}{diagnostic.message}
          </div>
          <div class="mt-1 opacity-80">{view.diagnosticEntityLabel(diagnostic)}</div>
        </div>
      {/each}
    </div>
  </div>
{/if}
