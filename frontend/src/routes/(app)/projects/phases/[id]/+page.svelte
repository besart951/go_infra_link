<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { ArrowLeft, Trash2 } from '@lucide/svelte';
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
  <div class="flex items-start gap-3">
    <Button variant="outline" onclick={() => goto('/projects/phases')}>
      <ArrowLeft class="mr-2 h-4 w-4" />
      {$t('common.back')}
    </Button>
    <div class="flex-1">
      <h1 class="text-3xl font-bold tracking-tight">
        {state.phase?.name ?? $t('phases.detail.fallback')}
      </h1>
      <p class="mt-1 text-muted-foreground">{$t('phases.detail.description')}</p>
    </div>
    {#if state.phase}
      <Button
        variant="destructive"
        size="sm"
        onclick={() => state.handleDelete()}
        disabled={state.busy}
      >
        <Trash2 class="mr-2 h-4 w-4" />
        {$t('common.delete')}
      </Button>
    {/if}
  </div>

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
