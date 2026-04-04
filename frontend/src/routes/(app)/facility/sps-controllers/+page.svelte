<script lang="ts">
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import SPSControllerListView from '$lib/components/facility/sps-controllers/SPSControllerListView.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.sps_controllers_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.sps_controllers_title')}</h1>
    <p class="text-sm text-muted-foreground">
      {$t('facility.sps_controllers_desc')}
    </p>
  </div>

  <SPSControllerListView />
</div>
