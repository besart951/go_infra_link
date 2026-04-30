<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useControlCabinetDetailState } from './state/context.svelte.js';

  const state = useControlCabinetDetailState();
  const t = createTranslator();

  function handleEditClick(): void {
    state.startCabinetEdit();
  }

  async function handleDeleteClick(): Promise<void> {
    await state.deleteCabinet();
  }
</script>

<EntityListHeader
  title={`${$t('facility.control_cabinet_detail.title')} #${state.cabinet.control_cabinet_nr}`}
  description={$t('facility.control_cabinet_detail.subtitle')}
  infoLabel={$t('common.info')}
  backHref="/facility/control-cabinets"
  backLabel={$t('common.back')}
>
    {#if state.canUpdateCabinet}
      <Button
        variant="outline"
        size="icon"
        onclick={handleEditClick}
        aria-label={$t('facility.control_cabinet_detail.edit_cabinet')}
      >
        <PencilIcon class="size-4" />
      </Button>
    {/if}
    {#if state.canDeleteCabinet}
      <Button
        variant="destructive"
        size="icon"
        onclick={handleDeleteClick}
        aria-label={$t('facility.control_cabinet_detail.delete_cabinet')}
      >
        <Trash2Icon class="size-4" />
      </Button>
    {/if}
</EntityListHeader>
