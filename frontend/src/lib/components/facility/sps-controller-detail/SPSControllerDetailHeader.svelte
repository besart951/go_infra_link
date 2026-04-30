<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
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

<EntityListHeader
  title={state.title}
  description={state.subtitle}
  infoLabel={$t('common.info')}
  backHref={state.backHref}
  backLabel={$t('common.back')}
>
    {#if state.canUpdateSps}
      <Button
        variant="outline"
        size="icon"
        onclick={handleEditClick}
        aria-label={$t('facility.sps_controller_detail.edit_controller')}
      >
        <PencilIcon class="size-4" />
      </Button>
    {/if}

    {#if state.canDeleteSps}
      <Button
        variant="destructive"
        size="icon"
        onclick={handleDeleteClick}
        aria-label={$t('facility.sps_controller_detail.delete_controller')}
      >
        <Trash2Icon class="size-4" />
      </Button>
    {/if}
</EntityListHeader>
