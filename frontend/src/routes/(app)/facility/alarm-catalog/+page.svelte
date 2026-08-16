<script lang="ts">
  import { onMount } from 'svelte';
  import { createTranslator } from '$lib/i18n/translator';
  import { AlarmCatalogState } from '$lib/components/facility/alarm-catalog/AlarmCatalogState.svelte.js';
  import AlarmCatalogUnitsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogUnitsSection.svelte';
  import AlarmCatalogFieldsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogFieldsSection.svelte';
  import AlarmCatalogTypesSection from '$lib/components/facility/alarm-catalog/AlarmCatalogTypesSection.svelte';
  import AlarmCatalogMappingsSection from '$lib/components/facility/alarm-catalog/AlarmCatalogMappingsSection.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';

  const t = createTranslator();
  const catalogState = new AlarmCatalogState({
    translate: (key, params) => $t(key, params)
  });
  let remoteChangePending = $state(false);

  onMount(() => {
    void catalogState.loadAll();

    return facilityReferenceDataCache.subscribeFacilityChanges((event) => {
      if (
        event.resource !== 'all' &&
        !['alarm_types', 'alarm_type_fields', 'alarm_fields', 'units'].includes(event.resource)
      ) {
        return;
      }
      if (catalogState.hasUnsavedChanges()) {
        if (!facilityReferenceDataCache.isChangeFromCurrentUser(event)) {
          remoteChangePending = true;
        }
        return;
      }
      void catalogState.loadAll();
    });
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
    backLabel={$t('common.back')}
  >
    {#if catalogState.loading}
      <span class="text-sm text-muted-foreground">
        {$t('facility.alarm_catalog_page.loading')}
      </span>
    {/if}
  </EntityListHeader>

  <div class="grid gap-6 xl:grid-cols-2">
    <AlarmCatalogUnitsSection state={catalogState} />
    <AlarmCatalogFieldsSection state={catalogState} />
  </div>

  {#if remoteChangePending}
    <div
      class="flex items-center justify-between gap-3 rounded-md border border-warning-border bg-warning-muted px-3 py-2 text-sm text-warning-muted-foreground"
      role="status"
    >
      <span>{$t('facility.realtime_change_pending')}</span>
      <button
        class="font-medium underline underline-offset-2"
        type="button"
        onclick={() => {
          remoteChangePending = false;
          void catalogState.loadAll();
        }}
      >
        {$t('common.refresh')}
      </button>
    </div>
  {/if}

  <div class="grid gap-6 xl:grid-cols-2">
    <AlarmCatalogTypesSection state={catalogState} />
    <AlarmCatalogMappingsSection state={catalogState} />
  </div>
</div>
