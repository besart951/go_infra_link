<script lang="ts">
  import type { PageData } from './$types.js';
  import { invalidateAll } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import HistoryIcon from '@lucide/svelte/icons/history';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import { createTranslator } from '$lib/i18n/translator.js';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
  import { SPSControllerSystemTypeDetailState } from '$lib/components/facility/sps-controller-system-type-detail/state/SPSControllerSystemTypeDetailState.svelte.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const detailState = new SPSControllerSystemTypeDetailState({
    data: () => data,
    invalidateAllAction: invalidateAll
  });
  let historyOpen = $state(false);

  async function handleRefreshAfterChange(): Promise<void> {
    await detailState.refreshAfterChange();
  }
</script>

<svelte:head>
  <title>{detailState.title} | {$t('facility.sps_controller')} | Infra Link</title>
</svelte:head>

<HistoryTimelineDialog
  bind:open={historyOpen}
  title={`${$t('history.title')}: ${detailState.title}`}
  scopeType="sps_controller_system_type"
  scopeId={data.systemType.id}
  controlCabinetId={data.cabinet?.id}
  onRestored={handleRefreshAfterChange}
/>

<div class="space-y-6">
  <EntityListHeader
    title={detailState.title}
    description={detailState.subtitle}
    backHref={detailState.backHref}
    backLabel={$t('common.back')}
  >
    <Button
      variant="outline"
      size="icon"
      onclick={() => (historyOpen = true)}
      aria-label={$t('history.open')}
    >
      <HistoryIcon class="size-4" />
    </Button>
    {#if detailState.canEdit}
      <Button
        variant="outline"
        size="icon"
        onclick={() => detailState.startEdit()}
        aria-label={$t('facility.sps_controller_system_type_detail.edit')}
      >
        <PencilIcon class="size-4" />
      </Button>
    {/if}
  </EntityListHeader>

  {#if detailState.showEdit}
    <SPSControllerForm
      initialData={data.controller}
      fixedControlCabinetId={data.controller.control_cabinet_id}
      onSuccess={handleRefreshAfterChange}
      onCancel={() => detailState.cancelEdit()}
    />
  {/if}

  <Card.Root class="border-border/70 bg-card/80">
    <Card.Header>
      <Card.Title class="text-xl">
        {$t('facility.sps_controller_system_type_detail.overview_title')}
      </Card.Title>
      <Card.Description>
        {$t('facility.sps_controller_system_type_detail.overview_desc')}
      </Card.Description>
    </Card.Header>

    <Card.Content>
      <div class="space-y-4">
        {#each detailState.overviewItems as item, index (item.label)}
          <div
            class={`grid gap-3 ${index < detailState.overviewItems.length - 1 ? 'border-b border-border/50 pb-4' : ''} sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]`}
          >
            <div class="text-xs font-medium tracking-[0.18em] text-muted-foreground uppercase">
              {item.label}
            </div>
            <div
              class={`text-sm font-medium text-foreground sm:text-right ${item.monospace ? 'font-mono' : ''}`}
            >
              {item.value}
            </div>
          </div>
        {/each}
      </div>
    </Card.Content>
  </Card.Root>

  <section class="space-y-3">
    <div class="space-y-1">
      <h2 class="text-xl font-semibold tracking-tight">
        {$t('facility.sps_controller_system_type_detail.field_devices_title')}
      </h2>
      <p class="text-sm text-muted-foreground">
        {$t('facility.sps_controller_system_type_detail.field_devices_desc')}
      </p>
    </div>

    {#key data.systemType.id}
      <FieldDeviceListView
        initialFilters={{
          buildingId: data.building?.id,
          controlCabinetId: data.cabinet?.id,
          spsControllerId: data.controller.id,
          spsControllerSystemTypeId: data.systemType.id
        }}
        systemTypeRefreshKey={data.systemType.id}
      />
    {/key}
  </section>
</div>
