<script lang="ts">
  import { onMount } from 'svelte';
  import { createTranslator } from '$lib/i18n/translator';
  import { AlarmCatalogState } from '$lib/components/facility/alarm-catalog/AlarmCatalogState.svelte.js';
  import AlarmCatalogUnitsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogUnitsSection.svelte';
  import AlarmCatalogFieldsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogFieldsSection.svelte';
  import AlarmCatalogTypesSection from '$lib/components/facility/alarm-catalog/AlarmCatalogTypesSection.svelte';
  import AlarmCatalogMappingsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogMappingsSection.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';

  const t = createTranslator();
  const state = new AlarmCatalogState({
    translate: (key) => $t(key)
  });

  onMount(() => {
    void state.loadAll();
  });
</script>

<svelte:head>
  <title>{$t('facility.alarm_catalog_page.title')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('facility.alarm_catalog_page.title')}
    description={$t('facility.alarm_catalog_page.description')}
    backHref="/facility"
    backLabel={$t('hub.back_to_overview')}
  >
    {#if state.loading}
      <span class="text-sm text-muted-foreground">
        {$t('facility.alarm_catalog_page.loading')}
      </span>
    {/if}
  </EntityListHeader>

  <div class="grid gap-6 xl:grid-cols-2">
    <AlarmCatalogUnitsSection {state} />
    <AlarmCatalogFieldsSection {state} />
  </div>

  <div class="grid gap-6 xl:grid-cols-2">
    <AlarmCatalogTypesSection {state} />
    <AlarmCatalogMappingsSection {state} />
  </div>
</div>
