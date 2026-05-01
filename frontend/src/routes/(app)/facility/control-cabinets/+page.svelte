<script lang="ts">
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import ControlCabinetListView from '$lib/components/facility/control-cabinets/ControlCabinetListView.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.control_cabinets_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('facility.control_cabinets_title')}
    description={$t('facility.control_cabinets_desc')}
    backHref="/facility"
    backLabel={$t('hub.back_to_overview')}
  />

  <ControlCabinetListView />
</div>
