<script lang="ts">
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator';
  import { canPerform } from '$lib/utils/permissions.js';

  const t = createTranslator();
</script>

<svelte:head>
  <title>{$t('facility.specifications')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('facility.specifications')}
    description={$t('facility.specifications_desc')}
    backHref="/facility"
    backLabel={$t('common.back')}
  />

  {#if canPerform('read', 'fielddevice')}
    <a href="/facility/field-devices" class="text-sm font-medium underline">
      {$t('facility.go_to_field_devices')}
    </a>
  {/if}
</div>
