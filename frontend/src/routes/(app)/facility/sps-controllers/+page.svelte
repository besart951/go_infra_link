<script lang="ts">
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import SPSControllerListView from '$lib/components/facility/sps-controllers/SPSControllerListView.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.sps_controllers_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('facility.sps_controllers_title')}
    description={$t('facility.sps_controllers_desc')}
    backHref="/facility"
    backLabel={$t('hub.back_to_overview')}
  />

  <SPSControllerListView />
</div>
