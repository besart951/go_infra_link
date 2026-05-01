<script lang="ts">
  import type { PageData } from './$types.js';
  import { invalidateAll } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import { createTranslator } from '$lib/i18n/translator.js';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import SPSControllerSystemTypeFieldDevicesCard from '$lib/components/facility/sps-controller-system-type-detail/SPSControllerSystemTypeFieldDevicesCard.svelte';
  import { SPSControllerSystemTypeDetailState } from '$lib/components/facility/sps-controller-system-type-detail/state/SPSControllerSystemTypeDetailState.svelte.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const state = new SPSControllerSystemTypeDetailState({
    data: () => data,
    invalidateAllAction: invalidateAll
  });

  async function handleRefreshAfterChange(): Promise<void> {
    await state.refreshAfterChange();
  }
</script>

<svelte:head>
  <title>{state.title} | {$t('facility.sps_controller')} | Infra Link</title>
</svelte:head>

<div class="space-y-6">
  <EntityListHeader
    title={state.title}
    description={state.subtitle}
    backHref={state.backHref}
    backLabel={$t('common.back')}
  >
    {#if state.canEdit}
      <Button
        variant="outline"
        size="icon"
        onclick={() => state.startEdit()}
        aria-label={$t('facility.sps_controller_system_type_detail.edit')}
      >
        <PencilIcon class="size-4" />
      </Button>
    {/if}
  </EntityListHeader>

  {#if state.showEdit}
    <SPSControllerForm
      initialData={data.controller}
      fixedControlCabinetId={data.controller.control_cabinet_id}
      onSuccess={handleRefreshAfterChange}
      onCancel={() => state.cancelEdit()}
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
        {#each state.overviewItems as item, index (item.label)}
          <div
            class={`grid gap-3 ${index < state.overviewItems.length - 1 ? 'border-b border-border/50 pb-4' : ''} sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]`}
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

  <SPSControllerSystemTypeFieldDevicesCard
    fieldDevices={state.fieldDevices}
    total={state.fieldDevicesTotal}
  />
</div>
