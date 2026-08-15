<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import ActivityEntryPoint from '$lib/components/activity/ActivityEntryPoint.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useControlCabinetDetailState } from './state/context.svelte.js';

  const detailState = useControlCabinetDetailState();
  const t = createTranslator();

  function handleEditClick(): void {
    detailState.startCabinetEdit();
  }

  async function handleDeleteClick(): Promise<void> {
    await detailState.deleteCabinet();
  }
</script>

<EntityListHeader
  title={`${$t('facility.control_cabinet_detail.title')} #${detailState.cabinet.control_cabinet_nr}`}
  description={$t('facility.control_cabinet_detail.subtitle')}
  backHref="/facility/control-cabinets"
  backLabel={$t('common.back')}
>
  <ActivityEntryPoint
    title={`${$t('history.title')}: ${detailState.cabinet.control_cabinet_nr ?? detailState.cabinet.id}`}
    scope={{ scopeType: 'control_cabinet', scopeId: detailState.cabinet.id }}
    controlCabinetId={detailState.cabinet.id}
    onRestored={() => detailState.refreshAfterChange()}
  />
  {#if detailState.canUpdateCabinet}
    <Button
      variant="outline"
      size="icon"
      onclick={handleEditClick}
      aria-label={$t('facility.control_cabinet_detail.edit_cabinet')}
    >
      <PencilIcon class="size-4" />
    </Button>
  {/if}
  {#if detailState.canDeleteCabinet}
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
