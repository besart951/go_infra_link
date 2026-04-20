<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import { useSPSControllerDetailState } from './state/context.svelte.js';

  const state = useSPSControllerDetailState();
  const t = createTranslator();

  function handleEditClick(): void {
    state.startEdit();
  }

  async function handleDeleteClick(): Promise<void> {
    await state.deleteController();
  }
</script>

<div class="flex flex-wrap items-start justify-between gap-4">
  <div class="flex items-start gap-3">
    <Button variant="ghost" size="icon" href={state.backHref} aria-label={$t('common.back')}>
      <ArrowLeftIcon class="size-4" />
    </Button>
    <div>
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">
        {state.title}
      </h1>
      <p class="text-sm text-muted-foreground">{state.subtitle}</p>
    </div>
  </div>

  <div class="flex items-center gap-2">
    {#if state.canUpdateSps}
      <Button variant="outline" size="sm" onclick={handleEditClick}>
        <PencilIcon class="mr-2 size-4" />
        {$t('facility.sps_controller_detail.edit_controller')}
      </Button>
    {/if}

    {#if state.canDeleteSps}
      <Button variant="destructive" size="sm" onclick={handleDeleteClick}>
        <Trash2Icon class="mr-2 size-4" />
        {$t('facility.sps_controller_detail.delete_controller')}
      </Button>
    {/if}
  </div>
</div>
