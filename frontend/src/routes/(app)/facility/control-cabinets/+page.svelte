<script lang="ts">
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import ControlCabinetListView from '$lib/components/facility/control-cabinets/ControlCabinetListView.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.control_cabinets_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.control_cabinets_title')}</h1>
    <p class="text-sm text-muted-foreground">{$t('facility.control_cabinets_desc')}</p>
  </div>

  <ControlCabinetListView />
</div>
