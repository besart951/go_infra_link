<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { Trash2 } from '@lucide/svelte';
  import { PhaseDetailPageState } from '$lib/components/project/PhaseDetailPageState.svelte.js';

  const t = createTranslator();

  const phaseId = $derived($page.params.id ?? '');
  const state = new PhaseDetailPageState(() => phaseId);

  onMount(() => {
    void state.load();
  });
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={state.phase?.name ?? $t('phases.detail.fallback')}
    description={$t('phases.detail.description')}
    infoLabel={$t('common.info')}
    backHref="/projects/phases"
    backLabel={$t('common.back')}
  >
    {#if state.phase}
      <Button
        variant="destructive"
        size="icon"
        onclick={() => state.handleDelete()}
        disabled={state.busy}
        aria-label={$t('common.delete')}
      >
        <Trash2 class="h-4 w-4" />
      </Button>
    {/if}
  </EntityListHeader>

  {#if state.error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="font-medium">{$t('phases.errors.load_title')}</p>
      <p class="text-sm">{state.error}</p>
    </div>
  {:else if state.loading}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      {$t('common.loading')}
    </div>
  {:else if state.phase}
    <div class="rounded-lg border bg-card p-6">
      <dl class="grid gap-4 text-sm">
        <div class="flex justify-between">
          <dt class="text-muted-foreground">Name</dt>
          <dd class="font-medium">{state.phase.name}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted-foreground">{$t('common.id')}</dt>
          <dd class="font-mono">{state.phase.id}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted-foreground">{$t('common.created')}</dt>
          <dd>{new Date(state.phase.created_at).toLocaleString()}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted-foreground">{$t('common.modified')}</dt>
          <dd>{new Date(state.phase.updated_at).toLocaleString()}</dd>
        </div>
      </dl>
    </div>
  {/if}
</div>
