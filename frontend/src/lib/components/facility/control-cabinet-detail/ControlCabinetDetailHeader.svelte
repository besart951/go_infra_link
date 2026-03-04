<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import type { ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    cabinet: ControlCabinet;
    canUpdate: boolean;
    canDelete: boolean;
    onEdit: () => void;
    onDelete: () => void;
  };

  let { cabinet, canUpdate, canDelete, onEdit, onDelete }: Props = $props();
  const t = createTranslator();
</script>

<div class="flex flex-wrap items-start justify-between gap-4">
  <div class="flex items-start gap-3">
    <Button variant="ghost" size="icon" href="/facility/control-cabinets" aria-label={$t('common.back')}>
      <ArrowLeftIcon class="size-4" />
    </Button>
    <div>
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">
        {$t('facility.control_cabinet_detail.title')} #{cabinet.control_cabinet_nr}
      </h1>
      <p class="text-sm text-muted-foreground">{$t('facility.control_cabinet_detail.subtitle')}</p>
    </div>
  </div>

  <div class="flex items-center gap-2">
    {#if canUpdate}
      <Button variant="outline" size="sm" onclick={onEdit}>
        <PencilIcon class="mr-2 size-4" />
        {$t('facility.control_cabinet_detail.edit_cabinet')}
      </Button>
    {/if}
    {#if canDelete}
      <Button variant="destructive" size="sm" onclick={onDelete}>
        <Trash2Icon class="mr-2 size-4" />
        {$t('facility.control_cabinet_detail.delete_cabinet')}
      </Button>
    {/if}
  </div>
</div>
