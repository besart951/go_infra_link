<script lang="ts">
  import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.field_devices_title')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('facility.field_devices_title')}
    description={$t('facility.field_devices_desc')}
    backHref="/facility"
    backLabel={$t('hub.back_to_overview')}
  />

  <FieldDeviceListView />
</div>
