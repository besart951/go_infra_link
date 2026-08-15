<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import ActivityEntryPoint from '$lib/components/activity/ActivityEntryPoint.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import { useSPSControllerDetailState } from './state/context.svelte.js';

  const detailState = useSPSControllerDetailState();
  const t = createTranslator();

  function handleEditClick(): void {
    detailState.startEdit();
  }

  async function handleDeleteClick(): Promise<void> {
    await detailState.deleteController();
  }
</script>

<EntityListHeader
  title={detailState.title}
  description={detailState.subtitle}
  backHref={detailState.backHref}
  backLabel={$t('common.back')}
>
  <ActivityEntryPoint
    title={`${$t('history.title')}: ${detailState.controller.device_name}`}
    scope={{ scopeType: 'sps_controller', scopeId: detailState.controller.id }}
    onRestored={() => detailState.refreshAfterChange()}
  />

  {#if detailState.canUpdateSps}
    <Button
      variant="outline"
      size="icon"
      onclick={handleEditClick}
      aria-label={$t('facility.sps_controller_detail.edit_controller')}
    >
      <PencilIcon class="size-4" />
    </Button>
  {/if}

  {#if detailState.canDeleteSps}
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
